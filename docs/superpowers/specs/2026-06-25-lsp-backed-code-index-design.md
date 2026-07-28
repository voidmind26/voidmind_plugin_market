# 基于 gopls 原生 MCP 的多仓代码索引增强设计

## 背景

`code-index-plugin` 已提供文件、Go 符号和代码块的本地静态索引，并曾在 `.mcp.json` 中独立注册 `gopls mcp`。健康 `gopls` 的符号、引用、类型、包 API、跨文件依赖和诊断比文本匹配可靠，因此应作为 Go 语义的首要信源。

独立注册只适用于任务工作目录本身是一个有效 Go workspace 的情况。当用户绑定一个包含多个独立仓库的父目录时，父目录通常没有 `go.work`，单个 `gopls` 进程不会递归把所有子目录的 `go.mod` 聚合为 workspace。此时它可能看到多个模块文件，却没有活动模块和包。

Codex 插件清单没有通用 `lspServers` 组件。能够声明的是实现 MCP 的 `gopls mcp`，而不是普通的 `gopls serve`。本次改造继续复用上游原生 MCP 语义，只在插件内部增加 MCP 到 MCP 的 workspace 路由。

## 目标

- 单仓、`go.work` 和多仓父目录使用同一组 `go_*` 工具。
- 按 `go.work` 或 `go.mod` 根目录隔离 `gopls mcp` 子会话。
- 工作区型工具显式接收 `project_root`，文件型工具从绝对路径确定 workspace。
- 父目录符号搜索和诊断受控分发到子 workspace，并合并结果。
- 子会话按根目录懒启动、并发复用，在 MCP 退出时统一关闭。
- 保持健康 `gopls` 的成功结果为高置信语义信源，同时保留静态索引和源码回退。

## 非目标

- 不重新实现 LSP framing、初始化、文档同步或语义查询。
- 不新增通用 LSP 插件清单字段。
- 不自动生成或修改用户项目的 `go.work`。
- 不自动安装或升级 `gopls`。
- 不默认对多仓父目录执行全量漏洞检查。
- 不注册 TypeScript、Python、Java 等其他语言服务器。

## 架构

插件只在 `.mcp.json` 注册一个 `code-index` stdio MCP：

```text
Codex
  |
  v
code-index MCP
  |-- 静态索引工具
  |-- list_go_workspaces
  `-- go_* 路由工具
        |
        |-- workspace A -> gopls mcp
        |-- workspace B -> gopls mcp
        `-- workspace C -> gopls mcp
```

路由层由四个职责组成：

- `workspace`：规范化路径，发现 `go.work` 或 `go.mod` 根目录。
- `session`：按 workspace root 懒启动、复用并关闭 `gopls mcp` 客户端。
- `router`：选择单仓或多仓调用方式，限制并发并聚合结果。
- `tools/gopls`：向 Codex 暴露稳定的 MCP 工具 schema。

## Workspace 解析

### 文件路径

文件型工具要求绝对路径。解析时向上查找：

1. 存在任意包含该文件的最近 `go.work` 时，使用该 `go.work` 根。
2. 否则使用最近的 `go.mod` 根。
3. 两者均不存在时返回可回退错误。

`go.work` 优先保证成员模块共享一个上游会话；没有 `go.work` 时，嵌套模块使用最近的 `go.mod`。

### 项目目录

如果 `project_root` 位于已有 workspace 内，直接使用包含它的 workspace。否则递归发现子目录中的 `go.work` 和 `go.mod`。发现一个 workspace 后不再扫描其内部模块，避免为同一构建范围重复启动会话。

扫描跳过版本控制目录、IDE 配置、`node_modules` 和 `vendor`，不跟随目录符号链接。所有会话键使用真实绝对路径，避免同一目录因符号链接产生重复进程。

## 工具路由

| 工具 | 路由依据 | 多仓父目录行为 |
|------|----------|----------------|
| `list_go_workspaces` | `project_root` | 仅发现并报告会话状态 |
| `go_workspace` | `project_root` | 分发并按 workspace 汇总 |
| `go_search` | `project_root` | 限并发分发，合并并去重符号行 |
| `go_file_context` | `file` | 自动选择文件所属 workspace |
| `go_package_api` | `project_root` | 要求明确单个 workspace |
| `go_symbol_references` | `file` | 自动选择文件所属 workspace |
| `go_diagnostics` | `files` 或 `project_root` | 文件按 workspace 分组；无文件时分发 |
| `go_vulncheck` | `project_root` | 要求明确单个 workspace |
| `go_rename_symbol` | `file` | 自动选择文件所属 workspace |

工作区型路由参数不会转发给上游，原生 gopls 参数名和语义保持不变。单仓调用直接返回上游结果；多仓调用按根目录稳定排序。部分子仓失败时保留成功结果并列出缺失范围，全部失败时才返回工具错误。

## 会话管理

第一次查询某个 workspace 时，通过登录 shell 在该根目录执行：

```bash
unset GOWORK
cd -- "$workspace_root"
exec gopls mcp
```

路由会清除继承的 `GOWORK` 覆盖，让 gopls 按已选择的 workspace 根自动识别 `go.work` 或 `go.mod`。根目录通过独立位置参数传入，不拼接到 shell 代码。客户端完成 MCP `initialize` 后才进入可复用状态；同一根目录的并发首次请求只会创建一个子进程。初始化失败的条目会被移除，后续调用允许重试。

服务持续消费子进程 stderr，避免管道写满阻塞；stderr 不作为语义证据。`code-index` MCP 收到 EOF、SIGINT 或 SIGTERM 退出时，统一关闭所有已创建的 gopls 客户端。

## 信源与回退

- `gopls`：Go 符号、引用、包 API、文件依赖和诊断的首要信源。
- 静态索引：候选文件、模块、标识符和非 Go 内容的发现信源。
- 源码读取：业务规则、控制流、动态行为和运行语义的解释信源。
- 文本搜索：动态注册、字符串引用和服务不可用时的回退信源。
- 测试与构建：代码变更正确性的最终验证。

工具缺失、`gopls` 版本不支持 MCP、目标没有 Go workspace 或子会话失败时，查询 skill 回退到静态索引与源码。多仓部分失败只降低对应 workspace 的结论覆盖范围。

## 安全边界

- 不自动执行 `go install` 或修改用户 Go 配置。
- 所有文件型调用要求绝对路径。
- 带 `project_root` 的诊断文件必须解析到其目录范围内的 workspace。
- `go_vulncheck.dir` 必须位于选定 workspace 内。
- `go_package_api` 和 `go_vulncheck` 不接受含多个独立 workspace 的父目录。
- `go_rename_symbol` 只返回编辑，不自动应用。

## 测试策略

自动测试覆盖：

- `go.work` 优先级、最近 `go.mod`、多仓发现和路径边界。
- 同一 workspace 的并发会话复用和退出关闭。
- 假上游 MCP 的请求代理、路由参数剥离、搜索合并去重。
- 跨仓诊断文件分组和单仓工具的歧义拒绝。
- `tools/list` 包含 4 个静态工具和 9 个 Go 路由工具。
- `.mcp.json` 只注册路由后的 `code-index` 服务。

真实 smoke test 使用可选父目录发现 workspace，再选择一个具体根启动真实 `gopls mcp`，验证 MCP 初始化、工具清单和 `go_workspace` 路由。多仓测试可通过环境变量声明预期 workspace 数量。

## 发布要求

- 同步两个插件清单、插件 README、根 README、`AGENTS.md` 与 `CLAUDE.md`。
- 运行 Go 单测、构建、skill 校验和 plugin 校验。
- 使用插件 cachebuster 脚本更新 Codex 版本并重新安装本地插件。
- 在新的 Codex 任务中验证更新后的 MCP 工具清单。
