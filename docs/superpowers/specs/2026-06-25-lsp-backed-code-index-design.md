# 基于 gopls 原生 MCP 的代码索引增强设计

## 背景

`code-index-plugin` 已提供文件、Go 符号和代码块的本地静态索引。静态索引适合快速发现候选，但无法可靠裁决符号定义、引用、类型、包 API、跨文件依赖和编译诊断。

Codex 插件规范允许声明 MCP server，但没有通用 `lspServers` 插件组件。任意语言服务器不能直接写入 `.mcp.json`，因为 LSP 和 MCP 虽然都使用 JSON-RPC 思想，初始化流程、消息协议和工具模型并不相同。

`gopls v0.22.0` 已提供原生 `gopls mcp` 子命令，能够通过 stdio 直接作为 MCP server 运行。因此 Go 语言第一版不再自行维护 LSP 到 MCP 的协议桥接，而是在代码索引插件中注册 `gopls` 的原生 MCP 服务。

## 目标

- 在 `code-index-plugin` 中直接注册 `gopls mcp`。
- 扩大 `gopls` 对 Go 代码查询、影响分析、修改和诊断的默认覆盖范围。
- 将健康 `gopls` 的成功结果视为高置信语义信源。
- 保留静态索引发现候选、非 Go 搜索和服务不可用时的回退能力。
- 保证 `gopls` 分析使用方工作区，而不是插件安装目录。
- 不自动安装或升级用户机器上的语言服务器。

## 非目标

- 不新增通用 LSP 插件清单字段。
- 不在自带 Go MCP 服务中实现 LSP Content-Length framing、会话路由或文档同步。
- 不连接 IDE 私有语言服务会话。
- 第一版不注册 TypeScript、Python、Java 等其他语言服务器。
- 不使用 hook 启动语言服务器、扫描项目或修改文件。

## 方案决策

### 采用方案：两个独立 MCP 服务

插件 `.mcp.json` 同时注册：

```text
code-index-plugin
├── code-index MCP
│   ├── build_code_index
│   ├── refresh_code_index
│   ├── search_code_index
│   └── get_code_index_status
└── gopls MCP
    ├── go_workspace
    ├── go_search
    ├── go_file_context
    ├── go_package_api
    ├── go_symbol_references
    ├── go_diagnostics
    ├── go_vulncheck
    └── go_rename_symbol
```

`code-index` 使用插件目录下的自带二进制，因此设置 `"cwd": "."`。`gopls` 必须继承当前使用方工作区，所以配置中不设置 `cwd`。

### 未采用方案：在 code-index MCP 内嵌 LSP 客户端

内嵌方案需要自行维护：

- LSP stdio 消息 framing 和 JSON-RPC 请求路由。
- `initialize`、文档打开与变更通知。
- 项目级 `gopls` 会话、并发、取消和退出清理。
- Location、LocationLink、DocumentSymbol、Hover 和诊断的协议兼容。
- MCP 工具 schema 与 LSP 结果归一化。

这些能力已由 `gopls mcp` 提供。重复实现会增加协议偏差、进程泄漏和版本兼容风险，且无法获得比上游更可靠的语义结果。

### 未采用方案：把 gopls 当作普通 LSP 写进 JSON

当前 Codex 插件清单没有 `lspServers` 或同类字段，未知字段会被插件校验拒绝。`.mcp.json` 也只能配置真正实现 MCP 的服务，不能直接启动 `gopls serve`。本设计能够直接声明，是因为使用的是 `gopls mcp`，不是 `gopls serve`。

## 启动与降级

`gopls` server 通过登录 shell 启动，以便读取用户安装 Go 工具时使用的 `PATH`：

```bash
gopls help mcp
exec gopls mcp
```

启动前检查：

1. `PATH` 中存在 `gopls`。
2. 当前版本支持 `mcp` 子命令。

缺少依赖或版本过低时，`gopls` MCP 启动失败并输出简短、可行动的 stderr。它与 `code-index` 是两个独立服务，因此静态索引能力不受影响。查询 skill 发现 `go_*` 工具不存在或 `go_workspace` 失败后，回退到静态索引和源码。

## 信源模型

- `gopls`：Go 符号、引用、包 API、文件内外依赖和诊断的首要信源。
- 静态索引：候选文件、模块、标识符、非 Go 内容和快速缩小范围的发现信源。
- 源码读取：业务规则、控制流、动态行为和运行语义的解释信源。
- 文本搜索：动态注册、字符串引用，以及 `gopls` 或索引不可用时的回退信源。
- 测试与构建：代码修改是否符合业务和运行要求的最终验证，不被语言服务替代。

不要求每个 `gopls` 成功结果都被文本搜索重复证明。语义结果与磁盘源码明显冲突时，先检查工作区、文件版本、符号选择和语言服务状态。

## 查询工作流

1. Go 任务且工具可用时，尽早调用 `go_workspace` 确认工作区。
2. 用户没有提供可靠文件或符号时，先用静态索引发现候选。
3. 用 `go_search` 确认符号位置，用 `go_symbol_references` 评估引用和影响。
4. 首次读取 Go 文件后，使用 `go_file_context` 获取同包依赖上下文。
5. 需要公开类型、签名或方法集合时，使用 `go_package_api`。
6. 需要错误、类型和静态分析结果时，使用 `go_diagnostics`。
7. 动态注册和字符串引用使用文本搜索补充，业务含义读取最小源码解释。

## 代码修改工作流

扩大 `gopls` 影响范围不只覆盖查询，也覆盖 Go 代码变更：

1. 修改包级符号前使用 `go_symbol_references` 检查影响范围。
2. 每批 Go 文件修改后，对活动文件调用 `go_diagnostics`。
3. 先解决 error 级诊断，再运行受影响包测试。
4. 新增或升级依赖后使用 `go_vulncheck`。
5. 只有用户明确要求重命名时才使用 `go_rename_symbol`，并审阅生成编辑后再应用。

## 安全与可移植性

- 不在仓库中写死本机 `gopls` 绝对路径。
- 不自动执行 `go install`。
- 不为 `gopls` 配置插件目录 `cwd`。
- `go_rename_symbol` 不作为普通查询工具使用。
- `gopls` stderr 只用于启动诊断，不作为语义证据。
- 静态索引和 `gopls` 均只面向当前使用方工作区。

## 测试策略

自动测试覆盖：

- `.mcp.json` 同时注册 `code-index` 和 `gopls`。
- `code-index` 固定在插件根目录启动。
- `gopls` 不设置 `cwd`，并执行 `gopls mcp`。
- 现有 4 个静态索引工具与全部 Go 测试继续通过。
- 查询 skill 只引用 `gopls` 实际公开的 `go_*` 工具。

真实 smoke test 覆盖：

- `gopls version` 和 `gopls help mcp` 成功。
- MCP `initialize` 与 `tools/list` 成功。
- 工具列表包含 8 个预期 `go_*` 工具。
- 在 Go 模块目录调用 `go_workspace` 能识别当前模块。

## 发布要求

- 同步 `.codex-plugin/plugin.json`、`.claude-plugin/plugin.json` 和市场版本。
- 同步插件 README、根 README、`AGENTS.md` 与 `CLAUDE.md`。
- 运行 skill 和 plugin 校验。
- 通过 cachebuster 生成新的 Codex 插件版本后提交发布。
