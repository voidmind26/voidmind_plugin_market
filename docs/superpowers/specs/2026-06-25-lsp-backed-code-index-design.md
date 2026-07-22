# 基于 LSP 的代码索引增强设计

## 背景

`code-index-plugin` 当前提供一个本地 MCP 服务，能力包括文件级索引、Go 符号索引、代码块索引、索引状态查询、刷新和搜索。Codex 插件规范没有把 LSP 作为一类可直接声明的插件组件，因此合规路径是把 LSP 放在插件 MCP 服务内部，由 MCP 工具对 Codex 暴露稳定能力。

本设计目标是在不依赖用户 IDE 会话的前提下，为 `code-index-plugin` 增加 LSP 驱动的代码理解能力。插件内部可以启动或连接语言服务器，但对 Codex 来说仍然只是一组 MCP 工具。语言服务状态健康且返回成功时，LSP 是符号语义事实的高置信信源，而不是仅供参考的弱提示。

## 目标

- 为 `code-index-plugin` 增加 LSP 导航和诊断能力。
- 使用静态索引发现候选位置，使用 LSP 优先确认符号语义，并保留静态回退能力。
- 增加 skill，引导 Codex 组合静态搜索、LSP 语义查询和最小源码读取。
- 仅在确实提升可发现性或一致性时增加轻量 hook。
- 保持插件可移植性，避免写死本机绝对路径。

## 信源模型

- 健康 LSP 的成功结果是定义、引用、类型、签名、文档符号和诊断的首要证据。
- 静态索引负责快速发现候选文件、模块和标识符，不与 LSP 竞争语义裁决权。
- 源码读取负责解释业务规则、控制流和运行语义，不要求用文本搜索机械重复证明每个 LSP 结果。
- 文本搜索负责动态注册、字符串引用以及 LSP/索引不可用时的回退。
- LSP 结果与工作区明显冲突时，先检查项目根目录、文件版本、目标 token、位置和语言服务状态，再决定是否降级。

## 非目标

- 不直接调用 IDE 私有的 LSP 会话。
- 不要求每个项目必须安装语言服务器后才能使用静态索引能力。
- 不实现完整编辑器、重构引擎或语义调用图。
- 不自动安装语言服务器。
- 不在 hook 中执行长时间扫描、构建或启动语言服务器。

## 架构

插件继续保持 Go MCP 服务形态。在现有 MCP 边界后增加内部 LSP 包：

```text
plugins/code-index-plugin/
├── internal/lsp/
│   ├── client/        # JSON-RPC / LSP 传输与生命周期
│   ├── registry/      # 语言服务器命令发现
│   ├── workspace/     # 项目根目录与文档 URI 辅助逻辑
│   └── service/       # 面向业务的 LSP 操作封装
├── internal/tools/index/
│   └── 在现有索引工具旁注册 LSP 工具
├── skills/
│   ├── code-index-search/
│   ├── code-index-init/
│   ├── code-index-refresh/
│   └── code-lsp-navigate/
└── hooks/             # 可选，仅用于注入轻量上下文提示
```

LSP 服务在 MCP 进程内按 `(project_root, language)` 管理语言服务器会话。会话在首次调用 LSP 工具时懒启动，带空闲超时，并在 MCP 进程退出时关闭。

## 支持语言

第一版支持范围应保守：

- Go：通过 `gopls`。
- TypeScript / JavaScript：仅在项目内可发现 `typescript-language-server` 或 `tsserver` 时启用。

如果找不到对应语言服务器，registry 必须返回清晰的不可用结果。静态索引搜索仍然必须可用。

## MCP 工具

在现有 `code-index` MCP 服务中新增以下工具：

- `get_lsp_status`
  - 输入：`project_root`，可选 `language`。
  - 输出：已配置语言、探测到的命令、运行中的会话、不可用原因。

- `lsp_document_symbols`
  - 输入：`project_root`、`path`。
  - 输出：符号树或扁平符号列表，包含范围信息。

- `lsp_definition`
  - 输入：`project_root`、`path`、`line`、`character`。
  - 输出：定义位置列表，包含文件路径、范围和来源标记。

- `lsp_references`
  - 输入：`project_root`、`path`、`line`、`character`，可选 `include_declaration`。
  - 输出：引用位置列表。

- `lsp_hover`
  - 输入：`project_root`、`path`、`line`、`character`。
  - 输出：简洁 hover 文本和范围。

- `lsp_diagnostics`
  - 输入：`project_root`，可选 `paths`。
  - 输出：按文件路径分组的诊断信息。

`line` 和 `character` 使用 LSP 的零基位置。工具描述中必须明确这一点。

## 数据流

1. Codex 先检查 LSP 工具和目标语言状态；用户已给出文件与位置时可直接进入语义查询。
2. 用户没有提供精确位置时，Codex 使用 `search_code_index` 缩小候选范围，再读取最小源码片段确定目标 token 和位置。
3. Codex 使用项目根目录、文件路径和零基位置调用与意图对应的 LSP 工具。
4. MCP handler 解析项目根目录和文件语言。
5. LSP registry 查找对应语言服务器命令。
6. LSP client 在需要时启动语言服务器，发送 `initialize`，打开文档，然后执行对应 LSP 方法。
7. 工具将 URI 和 range 结果转换为仓库相对路径和 JSON 友好的范围结构。
8. 成功结果作为高置信语义证据；如果 LSP 不可用或没有结果，工具返回结构化结果并回退到静态索引和源码读取。

## Skill 变更

更新现有 skill：

- `code-index-search`：使用静态索引发现候选；只要 LSP 工具真实可用且状态健康，就默认用 LSP 确认定义、引用、类型、符号结构和诊断。用户无需预先提供精确位置。
- `code-index-init`：说明 LSP 是可选增强，不替代静态索引构建。
- `code-index-refresh`：只刷新静态索引，不触发 LSP 会话。

新增 skill：

- `code-lsp-navigate`
  - 触发场景：用户询问定义、引用、hover / 类型信息、文档符号、诊断或语义导航。
  - 工作流：先调用 `get_lsp_status`；如果缺少位置，先用静态索引搜索定位候选文件并确定 token；再调用对应 LSP 工具；当 LSP 不可用时回退到源码读取或静态索引搜索。
  - 允许工具：新增 LSP 工具，以及现有 status / search 工具。

skill 应将健康 LSP 的成功结果视为高置信语义事实，同时不得在工具缺失、服务异常或工作区失配时伪造 LSP 结论。

## 可选 Hook

只有在确实需要提升可发现性时才增加 hook。hook 应满足：

- 在 `SessionStart` 运行。
- 注入一条简短上下文：`code-index-plugin` 使用索引发现候选，并在工具可用且状态健康时优先采用 LSP 语义结果。
- 不启动语言服务器，不扫描项目，不修改文件。

如果 hook 会增加市场安装噪音或信任确认成本，第一版应跳过 hook，先依赖 skill 发现。

## 错误处理

所有 LSP 工具返回结构化 JSON：

- `ok: true`，并携带结果数据。
- `ok: false`，`reason: "lsp_unavailable"`：找不到语言服务器。
- `ok: false`，`reason: "unsupported_language"`：文件语言暂不支持。
- `ok: false`，`reason: "position_invalid"`：行列位置非法。
- `ok: false`，`reason: "no_result"`：LSP 没有返回结果。
- 仅在请求格式错误或内部异常时返回 MCP tool error。

语言服务器 stderr 不应直接倾倒进正常结果，只保留简短、可行动的摘要。

## 配置

静态索引能力不需要项目配置。

未来可以支持按语言覆盖命令，但第一版应使用安全默认值和项目本地二进制：

- TypeScript / JavaScript 优先使用项目本地 `node_modules/.bin/typescript-language-server` 或 `tsserver`。
- Go 优先使用 `PATH` 中的 `gopls`。

## 测试

实现前必须按 TDD 推进。

初始集成测试覆盖：

- `tools/list` 包含新增 LSP 工具。
- 当找不到语言服务器时，`get_lsp_status` 返回不可用结果而不是失败。
- 使用 fake LSP server 验证 `lsp_definition`，并确认 MCP 工具能把 URI 转为路径。
- 现有静态索引工具继续通过。
- 查询 skill 在 LSP 可用时会为实现定位、定义和影响分析主动进入语义查询。
- 查询 skill 在 LSP 不可用或无结果时保留静态索引和源码回退。

聚焦单元测试覆盖：

- 按文件扩展名识别语言。
- 项目根目录与 URI 转换。
- 行列位置校验。
- LSP registry 命令发现。

手动 smoke check：

- 在 `plugins/code-index-plugin` 下执行 `GOWORK=off go test ./...`。
- 启动插件 MCP 并调用 `tools/list`。
- 在安装了 `gopls` 的 Go 项目中调用 `get_lsp_status` 和一个导航工具。

## 发布步骤

1. 增加 LSP service 接口和 fake-server 测试。
2. 增加 MCP 工具 schema 和 handler。
3. 增加最小 Go LSP 支持。
4. 如果实现成本可控，再增加 TypeScript / JavaScript 命令探测。
5. 更新 skills 和 README。
6. 只有在工具工作流稳定后，再决定是否增加 hook。

## 任务拆分概览

这是一个多单元功能，建议按行为层拆分：

1. LSP 核心与 MCP 工具：实现 service、registry、生命周期和工具 handler，并配套 fake-server 测试。
2. Skill 与文档更新：补充索引发现、LSP 语义优先和静态回退的使用流程。
3. 可选 hook：仅在前两部分稳定且仍确认有价值时实施。

依赖顺序是先核心 MCP 能力，再 skill / 文档，最后可选 hook。每个部分都可以单独形成一份实施计划。
