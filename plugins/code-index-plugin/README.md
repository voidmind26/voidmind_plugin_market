# code-index-plugin

为当前项目构建本地静态代码索引，并通过 `gopls` 原生 MCP 提供 Go 语义搜索、引用、包 API、文件上下文和诊断能力。

## 架构

插件在 `.mcp.json` 中注册两个独立的 stdio MCP 服务：

- `code-index`：插件自带的 Go 服务，负责索引构建、刷新、搜索和状态查询。
- `gopls`：执行 `gopls mcp`，直接使用语言服务器提供的 MCP 工具。

Codex 插件清单没有通用 `lspServers` 组件，因此不能把任意 LSP 进程当作 LSP 直接声明在 JSON 中。`gopls v0.22.0` 已原生实现 MCP stdio 服务，所以可以作为普通 MCP server 注册，不需要在 `code-index` 服务中重复实现 LSP JSON-RPC 桥接。

两个服务的工作目录有意不同：

- `code-index` 设置 `"cwd": "."`，从插件安装目录构建并启动自带二进制。
- `gopls` 不设置 `cwd`，继承使用方工作区，确保分析的是当前项目而不是插件源码。

任何修改都不得为 `gopls` 增加固定插件目录 `cwd`。

## 前置依赖

静态索引无需额外依赖。Go 语义能力要求 `PATH` 中存在支持 `mcp` 子命令的 `gopls`：

```bash
go install golang.org/x/tools/gopls@latest
gopls version
gopls help mcp
```

当前已验证版本为 `gopls v0.22.0`。插件不会自动安装或升级语言服务器；`gopls` 缺失或版本过低时，仅该 MCP 服务启动失败，静态索引仍可使用。

## 目录结构

```text
code-index-plugin/
├── .claude-plugin/plugin.json
├── .codex-plugin/plugin.json
├── .mcp.json
├── README.md
├── build.sh
├── cmd/code-index-mcp/main.go
├── go.mod
├── go.sum
├── plugin_config_test.go
├── internal/
│   ├── index/
│   │   ├── config/
│   │   ├── extractor/
│   │   ├── manifest/
│   │   ├── model/
│   │   ├── query/
│   │   ├── scanner/
│   │   ├── service/
│   │   └── storage/
│   ├── server/
│   └── tools/index/
└── skills/
    ├── code-index-init/
    ├── code-index-refresh/
    └── code-index-search/
```

## 静态索引工具

`code-index` 服务提供 4 个工具。

### `build_code_index`

构建当前项目的本地索引。

输入：

- `project_root`（可选）
- `deep_index_paths`（可选）

输出包含索引目录以及文件、符号和代码块数量。

### `refresh_code_index`

增量刷新已存在的索引，返回新增、变更、删除、未变文件数量及刷新后的统计。

### `search_code_index`

搜索当前项目索引。

输入：

- `query`
- `project_root`（可选）
- `path_prefix`（可选）
- `prefer_deep_hits`（可选）
- `limit`（可选，1 到 100）

结果包含命中类型、路径、展示行范围、摘要、评分和评分原因。

### `get_code_index_status`

查询索引是否就绪及当前统计。索引尚未建立时返回 `ready=false`，不直接报错。

## gopls 语义工具

`gopls` 服务原生提供 8 个 MCP 工具，具体 schema 由当前安装的 `gopls` 版本负责：

- `go_workspace`：识别 Go 模块、工作区和根目录。
- `go_search`：模糊搜索工作区 Go 符号。
- `go_file_context`：总结一个 Go 文件依赖的同包声明。
- `go_package_api`：查看一个或多个 Go 包的公开 API。
- `go_symbol_references`：查找包级符号、字段或方法的引用。
- `go_diagnostics`：获取工作区解析、构建和分析诊断。
- `go_vulncheck`：检查 Go 工作区依赖漏洞。
- `go_rename_symbol`：生成工作区符号重命名编辑。

读取类语义结果是定义、引用、类型、包 API、文件依赖和诊断的高置信信源。`go_rename_symbol` 只在用户明确要求重命名时使用，生成的编辑仍需审阅、应用和测试。

## 查询策略

`code-index-search` skill 按以下职责组合工具：

1. 静态索引快速发现候选文件、模块、标识符和非 Go 内容。
2. `gopls` 确认 Go 符号、引用、包 API、文件依赖和诊断。
3. 源码读取解释业务规则、控制流、动态注册和运行语义。
4. `gopls` 不可用或目标不是 Go 时，回退到静态索引、文本搜索和源码。

健康 `gopls` 的成功结果无需再用文本搜索机械复核。结果与磁盘源码明显冲突时，应先检查工作区、文件版本和符号选择，再决定是否降级。

## 索引存储

静态索引写入使用方项目：

```text
.claude/code-index/
├── manifest.json
├── files.jsonl
├── symbols.jsonl
└── chunks.jsonl
```

`gopls` 不读取或修改该索引目录，静态索引的构建与刷新也不会重启语言服务器。

## 本地验证

在插件目录执行：

```bash
GOWORK=off go test ./...
GOWORK=off go build ./...
gopls version
gopls help mcp
```

真实 MCP smoke test 应至少验证：

- `code-index` 的 `tools/list` 仍包含 4 个静态工具。
- `gopls` 的 `tools/list` 包含上述 8 个 `go_*` 工具。
- 在 Go 模块目录启动 `gopls mcp` 后，`go_workspace` 能识别当前模块。

## 当前边界

当前支持：

- 文件级索引
- Go 结构化符号索引
- 启发式代码块索引
- 索引构建、刷新、搜索和状态查询
- `gopls` 原生 MCP 语义与诊断工具
- 查询 skill 的 gopls 语义优先、静态发现与可靠回退策略

当前不做：

- 自动安装或升级 `gopls`
- 自建通用 LSP 协议桥接
- TypeScript、Python 等其他语言服务器注册
- 向量检索、远程索引服务或跨仓库联合索引
