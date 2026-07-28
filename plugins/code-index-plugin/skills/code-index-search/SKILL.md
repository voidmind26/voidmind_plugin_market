---
name: code-index-search
description: 用于定位代码结构、实现位置、符号定义与引用、调用影响、类型信息或诊断信息。Go 项目中优先使用插件路由的 gopls MCP 获取高置信语义事实；支持单仓、go.work 和包含多个独立 Go 仓库的父目录，静态索引用于发现候选与可靠回退，源码用于解释业务上下文。
---

# Code Index Search

组合静态代码索引与路由后的 `gopls mcp` 理解代码。健康 `gopls` 返回的符号、引用、包 API、文件依赖和诊断属于高置信语义事实；静态索引负责快速发现候选，源码负责解释业务规则与控制流。

## 能力边界

插件通过同一个 `code-index` MCP 服务提供：

- 静态索引：`get_code_index_status`、`build_code_index`、`refresh_code_index`、`search_code_index`。
- Go workspace 路由：`list_go_workspaces`。
- gopls 语义：`go_workspace`、`go_search`、`go_file_context`、`go_package_api`、`go_symbol_references`、`go_diagnostics`、`go_vulncheck`、`go_rename_symbol`。

路由层不重新实现 LSP 语义，只为每个 `go.work` 或 `go.mod` 根目录懒启动并复用独立的原生 `gopls mcp` 会话。只有当前会话真实提供这些工具时才能采用其结果；工具缺失、启动失败或工作区识别失败时，直接使用静态索引和源码回退，不要自行安装或升级语言服务器。

## 信源优先级

1. Go 符号、引用、包 API、跨文件依赖和诊断：优先采用健康 `gopls` 的成功结果。
2. 业务规则、控制流、动态注册和运行语义：读取相关源码并结合文本搜索解释。
3. 候选文件、模块、标识符和非 Go 内容发现：使用静态代码索引。
4. `gopls` 或索引不可用、无结果时：使用文本搜索和最小源码读取回退。

不要用文本搜索机械重复证明每个 `gopls` 结果。只有语义结果与当前磁盘源码明显冲突时，才检查路由到的 workspace、文件版本、符号名称和语言服务状态。

## 查询流程

### 1. 建立 Go workspace 上下文

当请求涉及 Go 代码时，先用当前任务的绝对项目目录调用 `list_go_workspaces(project_root)`：

- 返回一个 workspace：后续工作区型工具使用该根目录。
- 返回多个 workspace：这是多仓父目录；保留父目录用于 `go_search` 和无文件的 `go_diagnostics`，需要包 API 或漏洞检查时选择具体 workspace。
- 返回 `go.work`：把该 `go.work` 根作为一个 workspace，不再单独路由其成员模块。
- 没有 workspace：不要启动 `gopls`，转到静态索引和源码回退。

需要了解模块布局时调用 `go_workspace(project_root)`。多仓父目录会并发查询子仓并按 workspace 汇总；不要把单个子仓失败解释为其他仓库不存在目标代码。

### 2. 发现候选位置

用户已经给出可靠的 Go 文件或符号时，可直接进入语义查询。否则：

1. 调用 `get_code_index_status` 检查静态索引。
2. `ready=false` 时调用 `build_code_index`。
3. 使用简洁标识符、路由、配置键或领域词调用 `search_code_index`。
4. 查找具体函数、handler 或实现时设置 `prefer_deep_hits=true`；必要时用 `path_prefix` 缩小范围。
5. 对 Go 符号使用 `go_search(project_root, query)` 消除文本同名歧义。父目录搜索会查询各子 workspace 并合并去重；零结果不能证明业务逻辑不存在。

### 3. 用 gopls 确认语义

围绕用户问题选择最小充分工具集：

- 文件上下文：使用 `go_file_context(file)`，文件必须是绝对路径，路由层会选择最近的 `go.work` 或 `go.mod`。
- 包公开 API：使用 `go_package_api(project_root, packagePaths)`，`project_root` 必须指向单个 workspace，避免同名包歧义。
- 符号引用：使用 `go_symbol_references(file, symbol)`；动态注册和字符串引用再用文本搜索补充。
- 诊断：有活动文件时使用 `go_diagnostics(files)`，跨仓文件会自动分组；无活动文件时使用 `go_diagnostics(project_root)`，父目录会按子仓汇总。
- 漏洞检查：使用 `go_vulncheck(project_root, dir?, pattern?)`，必须明确单个 workspace，不对多仓父目录默认全量分发。
- 重命名：先用 `go_symbol_references` 评估影响，再调用 `go_rename_symbol(file, symbol, new_name)`；审阅编辑后才应用并验证。

所有文件路径使用绝对路径。不要无目的调用全部工具，也不要把 `go_search`、静态索引命中和源码解释混成同一种证据。

### 4. 扩大 gopls 对代码修改的影响

查询过程中若同时需要修改 Go 代码：

1. 修改符号定义前，先用 `go_symbol_references` 检查引用范围。
2. 每批 Go 文件修改后，对已编辑文件调用 `go_diagnostics(files)`；路由层会处理多仓分组。
3. 先修复 error 级诊断，再运行受影响包测试；提示级或信息级结果按任务相关性处理。
4. 新增或升级 Go 依赖后，对受影响的具体 workspace 使用 `go_vulncheck`。

`gopls` 是语义与诊断信源，不替代测试、业务验证或源码审查。

## 可靠回退

以下情况回退到静态索引、文本搜索和源码读取：

- 当前会话没有路由后的 `go_*` 工具。
- `gopls` 未安装、版本不支持 `mcp` 子命令或子会话启动失败。
- 目标路径下没有 `go.work` 或 `go.mod`。
- 目标不是 Go 代码。
- 语义查询没有结果，或结果与当前磁盘源码明显不一致。

多仓分发返回部分成功时，继续使用成功 workspace 的结果，并明确列出未完成的 workspace。不要把局部启动失败扩大为整个父目录不可用。

## 输出要求

- 先给最相关结论，再列文件路径和精确行号。
- 将成功的 `go_*` 结果标为“gopls 语义确认”，并与“静态索引命中”“源码解释”区分。
- 多仓结果注明命中的 workspace；部分失败时说明缺失范围。
- 定义或引用查询只保留关键位置和必要上下文，不倾倒全部结果。
- `gopls` 不可用时只说明可行动的回退原因，不原样输出子进程 stderr。
