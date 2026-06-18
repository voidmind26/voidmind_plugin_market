# voidmind_plugin_market

私有 zcode / Claude Code 插件市场仓库，用于集中管理多个内部插件。

## 目录结构

```text
├── .claude-plugin/
│   └── marketplace.json       # 市场级插件清单，注册所有插件来源
├── plugins/                   # 插件目录，每个子目录为一个独立插件
│   ├── apifox-codegen-plugin/      # Apifox 接口文档与测试用例生成
│   ├── backend-construct-plugin/   # 后端 plan 统一入口
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
| [apifox-codegen-plugin](plugins/apifox-codegen-plugin/) | 0.1.1 | 从 Go Web 代码生成 Apifox 接口文档与场景化测试用例 |
| [backend-construct-plugin](plugins/backend-construct-plugin/) | 0.2.2 | 后端 plan 阶段统一入口，提供标签驱动计划与复杂任务辅助 |
| [betterpowers](plugins/betterpowers/) | 5.1.5 | 优化版 superpowers，包含多项开发技能、钩子与工作流 |
| [code-index-plugin](plugins/code-index-plugin/) | 0.1.2 | 为当前项目构建本地代码索引并提供搜索、刷新、状态查询的 MCP 工具 |
| [fusion-mcp](plugins/fusion-mcp/) | 0.1.0 | Autodesk Fusion 360 MCP 集成 |
| [gateway-platform-plugin](plugins/gateway-platform-plugin/) | 1.0.1 | 本地 HTTP MCP 网关平台，提供 route、secret、动态注入与 Web Console |
| [local-db-access](plugins/local-db-access/) | 0.3.3 | 本地数据库访问 MCP 插件，提供查询、受限写入和连接配置初始化能力 |

## 添加新插件

1. 复制 `plugins/plugin-template` 到 `plugins/<plugin-name>/`
2. 修改 `plugins/<plugin-name>/.claude-plugin/plugin.json`
3. 按需创建 `commands/`、`agents/`、`skills/`、`hooks/` 或 `.mcp.json`
4. 更新 `.claude-plugin/marketplace.json` 注册插件来源
5. 更新本 `README.md` 的插件列表

## 插件规范

每个插件需包含 `.claude-plugin/plugin.json`，遵循标准 zcode 插件结构。插件内部引用路径优先使用 `${CLAUDE_PLUGIN_ROOT}`，避免硬编码本机绝对路径。
