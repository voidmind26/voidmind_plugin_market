# code-index-plugin

为当前项目构建本地静态代码索引，并通过多 workspace `gopls mcp` 路由提供 Go 语义搜索、引用、包 API、文件上下文和诊断能力。

## 架构

插件在 `.mcp.json` 中只注册一个 `code-index` stdio MCP 服务。该服务同时提供：

- 4 个静态索引工具，负责索引构建、刷新、搜索和状态查询。
- 1 个 Go workspace 发现工具。
- 8 个与原生 `gopls mcp` 同名的 `go_*` 语义工具。

`code-index` 固定从插件根目录启动，因此所有工作区型 Go 工具都显式接收 `project_root`。服务根据路径发现 `go.work` 或 `go.mod`，为每个 workspace 懒启动并复用独立的 `gopls mcp` 子会话。

```text
Codex -> code-index MCP -> workspace router -> gopls mcp per workspace
                         `-> static code index
```

Codex 插件清单没有通用 `lspServers` 组件，不能把任意 LSP 进程直接声明在 JSON 中。这里复用的是 `gopls v0.22.0+` 原生 MCP 服务；路由层只负责 workspace 选择和结果聚合，不重复实现 LSP。

## 多仓行为

当绑定目录本身是 Go module 或 `go.work` 时，路由使用该 workspace。当绑定目录是没有 `go.work` 的多仓父目录时，路由递归发现各子目录的 `go.mod`：

- `go_workspace`、`go_search` 和无文件参数的 `go_diagnostics` 可分发到多个子仓。
- `go_search` 合并并去重结果，分发并发上限为 4。
- `go_diagnostics(files)` 按文件所属 workspace 自动分组。
- `go_package_api` 和 `go_vulncheck` 要求 `project_root` 指向具体单仓，避免包名歧义和意外全量扫描。
- `go_file_context`、`go_symbol_references` 和 `go_rename_symbol` 从绝对文件路径选择 workspace。

路由优先使用包含目标路径的 `go.work`；不存在 `go.work` 时使用最近的 `go.mod`。一个子仓启动失败不会丢弃其他子仓的成功结果。

启动子会话时会清除继承的 `GOWORK` 覆盖，确保 gopls 按路由选定的根自动识别对应 `go.work` 或 `go.mod`。

## 前置依赖

静态索引无需额外依赖。Go 语义能力要求 `PATH` 中存在支持 `mcp` 子命令的 `gopls`：

```bash
go install golang.org/x/tools/gopls@latest
gopls version
gopls help mcp
```

当前已验证版本为 `gopls v0.22.0`。插件不会自动安装或升级语言服务器；`gopls` 缺失或版本过低时，Go 路由工具返回可行动错误，静态索引仍可使用。

## 目录结构

```text
code-index-plugin/
├── .claude-plugin/plugin.json
├── .codex-plugin/plugin.json
├── .mcp.json
├── README.md
├── build.sh
├── cmd/code-index-mcp/main.go
├── internal/
│   ├── gopls/
│   │   ├── router/
│   │   ├── session/
│   │   └── workspace/
│   ├── index/
│   ├── server/
│   └── tools/
│       ├── gopls/
│       └── index/
└── skills/
    ├── code-index-init/
    ├── code-index-refresh/
    └── code-index-search/
```

## 静态索引工具

### `build_code_index`

使用 `project_root` 和可选的 `deep_index_paths` 构建本地索引，返回索引目录以及文件、符号和代码块数量。

### `refresh_code_index`

增量刷新已有索引，返回新增、变更、删除、未变文件数量和刷新后统计。

### `search_code_index`

按 `query` 搜索索引，可使用 `project_root`、`path_prefix`、`prefer_deep_hits` 和 1 到 100 的 `limit` 缩小结果。

### `get_code_index_status`

查询索引是否就绪及当前统计；索引尚未建立时返回 `ready=false`。

## Go workspace 与语义工具

### `list_go_workspaces`

输入 `project_root`，返回发现的 workspace 根、标记类型、标记文件和当前子会话是否已经启动。该工具只发现目录，不启动 `gopls`。

### 路由后的原生工具

- `go_workspace(project_root)`：汇总一个或多个 Go workspace。
- `go_search(project_root, query)`：模糊搜索一个或多个 workspace 的 Go 符号。
- `go_file_context(file)`：总结 Go 文件依赖的同包声明。
- `go_package_api(project_root, packagePaths)`：查看单个 workspace 中包的公开 API。
- `go_symbol_references(file, symbol)`：查找包级符号、字段或方法的引用。
- `go_diagnostics(project_root?, files?)`：获取解析、构建和静态分析诊断。
- `go_vulncheck(project_root, dir?, pattern?)`：检查单个 workspace 的依赖漏洞。
- `go_rename_symbol(file, symbol, new_name)`：生成 workspace 符号重命名编辑。

读取类结果是 Go 定义、引用、类型、包 API、文件依赖和诊断的高置信信源。`go_rename_symbol` 不自动应用编辑，仍需审阅和测试。

## 查询策略

`code-index-search` skill 按以下职责组合工具：

1. 先用 `list_go_workspaces` 明确单仓、`go.work` 或多仓父目录。
2. 静态索引快速发现候选文件、模块、标识符和非 Go 内容。
3. 路由后的 `gopls` 确认 Go 符号、引用、包 API、文件依赖和诊断。
4. 源码读取解释业务规则、控制流、动态注册和运行语义。
5. `gopls` 不可用或目标不是 Go 时，回退到静态索引、文本搜索和源码。

## 索引存储

静态索引写入使用方项目：

```text
.claude/code-index/
├── manifest.json
├── files.jsonl
├── symbols.jsonl
└── chunks.jsonl
```

`gopls` 子会话不读取或修改该目录，静态索引构建与刷新也不会重启语言服务器。

## 本地验证

```bash
GOWORK=off go test ./...
GOWORK=off go build ./...
gopls version
gopls help mcp
```

真实单仓 smoke test：

```bash
CODE_INDEX_GOPLS_SMOKE=1 GOWORK=off go test -run TestGoplsRouterSmoke -v .
```

增加 `CODE_INDEX_GOPLS_SMOKE_BINARY=1` 可让 smoke test 使用 `build.sh` 生成的 `bin/code-index-mcp`，验证发布二进制而不是 `go run`。

真实多仓父目录 smoke test：

```bash
CODE_INDEX_GOPLS_SMOKE=1 \
CODE_INDEX_GOPLS_SMOKE_ROOT=/path/to/multi-repo-parent \
CODE_INDEX_GOPLS_SMOKE_WORKSPACES=9 \
CODE_INDEX_GOPLS_SMOKE_FANOUT=1 \
GOWORK=off go test -run TestGoplsRouterSmoke -v .
```

真实 `go.work` 路由 smoke test：

```bash
CODE_INDEX_GOPLS_SMOKE=1 \
CODE_INDEX_GOPLS_SMOKE_GOWORK=1 \
GOWORK=off go test -run TestGoplsRouterSmoke -v .
```

## 当前边界

当前支持文件级索引、Go 结构化符号索引、启发式代码块索引、索引构建/刷新/搜索，以及单仓和多仓的 `gopls mcp` 语义路由。

当前不自动安装或升级 `gopls`，不生成 `go.work`，不实现通用 LSP 桥接，不注册其他语言服务器，也不提供向量检索或远程索引服务。
