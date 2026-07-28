# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

私有 zcode / Claude Code 插件市场仓库，用于集中管理多个内部插件。根级市场清单在 `.claude-plugin/marketplace.json`，各插件在 `plugins/<plugin-name>/` 下独立维护自己的 `.claude-plugin/plugin.json`、技能、代理、钩子或 MCP 配置。

## 仓库结构

```text
├── .claude-plugin/marketplace.json   # 市场级清单，注册所有插件来源
├── plugins/
│   └── <plugin-name>/
│       ├── .claude-plugin/plugin.json # 单插件清单
│       ├── commands/                  # slash command 组件（可选）
│       ├── agents/                    # 子代理组件（可选）
│       ├── skills/<skill>/SKILL.md    # skill 组件（可选）
│       ├── hooks/hooks.json           # hook 组件（可选）
│       └── .mcp.json                  # MCP 服务配置（可选）
├── CLAUDE.md
└── README.md
```

## 架构速览

- 根级 `.claude-plugin/marketplace.json` 是插件市场注册源，新增、删除或升级插件时同步维护这里和 `README.md` 的插件列表。
- 每个 `plugins/<plugin-name>/` 是独立插件边界；插件内部引用路径优先使用 `${CLAUDE_PLUGIN_ROOT}`，不要写死本机绝对路径。
- `apifox-codegen-plugin`、`backend-construct-plugin`、`betterpowers` 主要提供 skills/agents/hooks 等 Claude Code 行为组件；修改 skill 内容时优先保持触发描述、工作流边界和参考文档一致。`backend-construct-plugin` 是独立的后端代码规范与实现插件，不依赖 `betterpowers`。
- `code-index-plugin` 只注册自带的 Go MCP 服务，同时提供静态索引和多 workspace `gopls mcp` 路由。路由按 `project_root` 或绝对文件路径发现 `go.work`/`go.mod`，为每个 workspace 懒启动独立 gopls 子会话；索引数据写入使用方项目的 `.claude/code-index/`。
- `gateway-platform-plugin` 是复合插件：Go MCP/HTTP 服务负责本地网关与 SQLite 数据，`frontend/` 是 Vue Web Console，构建产物会复制到 `server/router/frontend_dist/` 供 Go 服务内嵌。
- `fusion-mcp` 通过 `.mcp.json` 接入 Fusion 360 MCP 能力，当前主要是外部 MCP 集成配置。

## 添加或调整插件

- 新插件优先复制 `plugins/plugin-template` 作为起点，然后修改 `plugins/<plugin-name>/.claude-plugin/plugin.json`。
- 只创建实际需要的 `commands/`、`agents/`、`skills/`、`hooks/` 等目录，不要为了结构完整创建空目录。
- 目录、文件和组件名统一使用 `kebab-case`。
- 新增插件后同步更新 `.claude-plugin/marketplace.json` 和 `README.md`；若插件暴露 MCP 服务，同时维护插件内 `.mcp.json`。

## 已注册插件速览

| 插件 | 主要组件 |
|------|----------|
| `plugins/apifox-codegen-plugin/` | skills: `apifox-dev`, `generate-interfaces-from-code`, `generate-scenario-tests`; Apifox HTTP MCP 配置 |
| `plugins/backend-construct-plugin/` | skill: `backend-dev`; 后端代码规范资产：`knowledge/`、`references/`、`examples/` |
| `plugins/betterpowers/` | hooks; 通用开发流程 skills，如 `brainstorming`、`test-driven-development`、`systematic-debugging`、`writing-plans` 等 |
| `plugins/code-index-plugin/` | 静态索引 MCP + 多 workspace gopls MCP 路由; skills: `code-index-init`, `code-index-refresh`, `code-index-search` |
| `plugins/fusion-mcp/` | Fusion 360 MCP 集成配置 |
| `plugins/gateway-platform-plugin/` | Go MCP/HTTP 网关、SQLite 数据层、Vue Web Console、Python 测试 |

`plugins/betterpowers/` 有独立的 `CLAUDE.md` 和 `hooks/hooks.json`；修改该目录时必须遵守其子目录指南。
