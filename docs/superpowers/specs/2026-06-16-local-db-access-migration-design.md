# local-db-access 插件迁移设计

## 背景

当前 `zyb-test-db-plugin` 位于 `aitools/plugins/zyb-test-db-plugin/`，目标是迁移到 `voidmind_plugin_market/plugins/local-db-access/`，并以 `local-db-access` 作为新插件名称注册到目标插件市场。

迁移范围是插件品牌、目录、元数据、MCP 服务命名、Go 模块路径、skill 工具引用和目标市场清单。数据库查询、受限写入、安全边界、配置格式和 MCP 工具业务能力保持不变。

## 目标

- 在目标仓库创建 `plugins/local-db-access/`，内容来自源插件。
- 插件元数据名称改为 `local-db-access`，描述调整为本地数据库访问插件，不再使用 `zyb-test-db-plugin` 旧品牌。
- 目标市场根 `.claude-plugin/marketplace.json` 注册 `local-db-access`，并同步更新目标仓库 `README.md`。
- 插件内部所有会影响构建、工具暴露或用户文档的旧名称同步替换。
- 保持 MCP 工具能力和数据库访问策略不变：只读查询仍仅允许查询类 SQL，受限写入仍仅允许 `INSERT` / `UPDATE` / `CREATE` / `ALTER`，继续拒绝 `DELETE` / `DROP` / `TRUNCATE` 与多语句拼接。

## 非目标

- 不修改源仓库 `aitools/plugins/zyb-test-db-plugin/`。
- 不调整数据库连接字段、配置 YAML schema 或安全策略。
- 不引入项目级配置覆盖机制。
- 不改造 MCP 工具业务逻辑和 SQL 校验规则。
- 不提交 git commit，除非用户后续明确要求。

## 设计决策

### 推荐方案：完整品牌迁移

复制源插件目录到目标仓库后，系统性替换插件名称和关联标识：

- 插件目录：`zyb-test-db-plugin` → `local-db-access`
- 插件清单名：`zyb-test-db-plugin` → `local-db-access`
- Go module：`zyb-test-db-plugin` → `local-db-access`
- Go import：`zyb-test-db-plugin/internal/...` → `local-db-access/internal/...`
- MCP 服务名：`test-db` → `local-db`
- 二进制名：`test-db-mcp` → `local-db-access-mcp`
- 用户配置/文档中的旧插件标识：`zyb-test-db-plugin` → `local-db-access`
- skill allowed-tools：`mcp__plugin_zyb-test-db-plugin_test-db__...` → `mcp__plugin_local-db-access_local-db__...`

该方案让目录名、插件名、MCP 服务名和工具命名一致，避免目标市场中出现新旧品牌混用。

### 备选方案：仅改插件目录和清单

只复制目录并修改 `plugin.json` 与市场清单，保留 Go module、MCP 服务名、二进制名和 skill 工具引用。该方案改动少，但安装后用户看到的工具名仍包含旧插件和旧服务名，容易误导，也不符合“修改名称为 local-db-access”的目标，因此不采用。

## 架构与组件

### 插件目录

目标目录为：

```text
plugins/local-db-access/
├── .claude-plugin/plugin.json
├── .mcp.json
├── README.md
├── build.sh
├── cmd/test-db-mcp/main.go
├── go.mod
├── go.sum
├── internal/
└── skills/
```

`cmd/test-db-mcp/` 目录可以保留，以减少文件移动；对外暴露的二进制改为 `bin/local-db-access-mcp`。如果后续需要进一步清理，可单独把入口目录改为 `cmd/local-db-access-mcp/`，但本次迁移不做额外重构。

### MCP 服务

`.mcp.json` 注册服务名 `local-db`：

- command 仍使用 `bash -lc`。
- 启动时优先执行 `${CLAUDE_PLUGIN_ROOT}/bin/local-db-access-mcp`。
- 二进制不存在时在插件根目录执行 `GOWORK=off go build -o "$BIN" ./cmd/test-db-mcp`。
- 使用 `${CLAUDE_PLUGIN_ROOT}`，不写死本机绝对路径。

### Go 模块

`go.mod` module 改为 `local-db-access`，所有内部 import 同步替换。业务代码行为保持不变。

### Skills

保留现有 skill 名称：

- `test-db-query`
- `test-db-write`
- `init-test-db-config`

原因是这些 skill 名称描述的是用户意图，不是插件品牌。skill 内容中的插件名、工具 allowed-tools 和说明文字同步改为 `local-db-access` / `local-db`。

### 文档

`plugins/local-db-access/README.md` 更新：

- 标题改为 `local-db-access`。
- 目录路径和 zcode 示例改为目标仓库路径。
- MCP 服务名改为 `local-db`。
- 二进制名改为 `local-db-access-mcp`。
- 内部配置路径说明改为 `.claude/local-db-access.local.md` 或删除旧的项目级配置描述，以匹配当前 skill 已声明的“仅插件内部配置”。

目标仓库 `README.md` 更新：

- 目录结构加入 `local-db-access`。
- 已注册插件表加入 `local-db-access` 版本和说明。

目标仓库 `.claude-plugin/marketplace.json` 更新：

- 新增插件项：`name` 为 `local-db-access`，`source` 为 `./plugins/local-db-access`。
- `version` 继承源插件 `0.3.3`。
- `author` 继承源插件作者。

## 数据流

1. Claude Code / zcode 从目标市场安装或加载 `local-db-access`。
2. 插件通过 `.mcp.json` 启动 `local-db` MCP 服务。
3. 启动脚本在 `${CLAUDE_PLUGIN_ROOT}` 下查找或构建 `bin/local-db-access-mcp`。
4. Go 服务读取 `${CLAUDE_PLUGIN_ROOT}/internal/config/config.yaml`。
5. skill 通过新工具名调用 MCP 工具。
6. MCP 工具继续使用现有数据库管理器执行只读查询、受限写入、表结构读取、健康检查和配置初始化。

## 错误处理

- 如果 `CLAUDE_PLUGIN_ROOT` 未设置，沿用现有错误返回，提示环境变量缺失。
- 如果二进制不存在，启动脚本自动构建；构建失败时 MCP 启动失败并暴露 Go 编译错误。
- 如果配置 YAML 缺失或非法，沿用现有配置读取和校验错误。
- 如果 skill allowed-tools 未同步，工具会因名称不匹配不可用；迁移验证必须覆盖该项。

## 测试与验证

迁移后在目标插件目录执行：

```bash
GOWORK=off go test ./...
GOWORK=off go build -o bin/local-db-access-mcp ./cmd/test-db-mcp
```

再从目标仓库或任意项目用插件目录做最小 smoke test：

```bash
zcode run --plugin-dir /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access \
  --allowedTools "mcp__plugin_local-db-access_local-db__list_databases" \
  -p "先列出有哪些测试库"
```

如果 smoke test 需要真实数据库配置而当前环境没有可用连接，可以只完成 Go test/build，并在最终说明中标记 zcode smoke test 未执行或受配置限制。

## 任务拆解概览

单一可计划单元；后续 plan skill 将其拆成复制目录、重命名标识、更新市场清单、更新文档和验证构建测试等步骤。

## 审核重点

- `local-db-access` 是否应同时替换 MCP 服务名为 `local-db`。
- 是否接受保留 `cmd/test-db-mcp/` 入口目录以减少迁移重构。
- 是否接受 skill 名称仍保持 `test-db-query`、`test-db-write`、`init-test-db-config`，只更新其内部工具引用和说明。
