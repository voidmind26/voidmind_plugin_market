# voidmind_plugin_market

私有 zcode / Claude Code 插件市场仓库，用于集中管理多个内部插件。

## 目录结构

```text
├── .claude-plugin/
│   └── marketplace.json       # 旧版 zcode / Claude 插件市场清单
├── .agents/
│   └── plugins/marketplace.json # Codex repo-local 插件市场清单
├── plugins/                   # 插件目录，每个子目录为一个独立插件
│   ├── apifox-codegen-plugin/      # 基于 Apifox CLI 的接口与测试生成
│   ├── backend-construct-plugin/   # Go 后端代码规范与实现
│   ├── betterpowers/               # 优化版 superpowers 工作流技能
│   ├── code-index-plugin/          # 本地代码索引 MCP 插件
│   ├── fusion-mcp/                 # Fusion 360 MCP 集成
│   ├── gateway-platform-plugin/    # 本地 HTTP MCP 网关与 Web Console
│   ├── local-db-access/            # 本地数据库访问 MCP 插件
│   └── plugin-template/            # 插件模板（新建插件时复制）
├── CLAUDE.md
└── README.md
```

## 已注册插件

| 插件 | 版本 | 说明 |
|------|------|------|
| [apifox-codegen-plugin](plugins/apifox-codegen-plugin/) | 0.1.1 | 通过官方 Apifox CLI 从 Go Web 代码生成接口文档与场景化测试用例 |
| [backend-construct-plugin](plugins/backend-construct-plugin/) | 0.3.0 | 按项目约定指导 Go 后端代码生成、修改与验证 |
| [betterpowers](plugins/betterpowers/) | 5.1.5 | 优化版 superpowers，包含多项开发技能、钩子与工作流 |
| [code-index-plugin](plugins/code-index-plugin/) | 0.2.0 | 使用静态索引发现候选，并通过多 workspace gopls MCP 路由确认 Go 语义 |
| [fusion-mcp](plugins/fusion-mcp/) | 0.1.0 | Autodesk Fusion 360 MCP 集成 |
| [gateway-platform-plugin](plugins/gateway-platform-plugin/) | 1.0.1 | 本地 HTTP MCP 网关平台，提供 route、secret、动态注入与 Web Console |
| [local-db-access](plugins/local-db-access/) | 0.3.3 | 本地数据库访问 MCP 插件，提供查询、受限写入和连接配置初始化能力 |

## 添加新插件

1. 复制 `plugins/plugin-template` 到 `plugins/<plugin-name>/`
2. 修改 `plugins/<plugin-name>/.claude-plugin/plugin.json` 和 `plugins/<plugin-name>/.codex-plugin/plugin.json`
3. 按需创建 `commands/`、`agents/`、`skills/`、`hooks/` 或 `.mcp.json`
4. 更新 `.claude-plugin/marketplace.json` 和 `.agents/plugins/marketplace.json` 注册插件来源
5. 更新本 `README.md` 的插件列表

## 插件规范

每个插件保留 `.claude-plugin/plugin.json` 兼容旧版 zcode / Claude 插件结构，同时新增 `.codex-plugin/plugin.json` 供 Codex Desktop 识别。

使用插件内相对路径启动二进制的 `stdio` MCP，应显式设置 `"cwd": "."`，避免依赖 `${CLAUDE_PLUGIN_ROOT}`。需要分析使用方工作区的子进程必须显式接收工作区路径；例如 `code-index-plugin` 的路由层按 `project_root` 在具体 workspace 根启动 `gopls mcp`。HTTP MCP 插件可继续使用固定本地或远端 URL。
