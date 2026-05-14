# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

私有 Claude Code 插件集合仓库（插件市场），用于存放和管理自定义 zcode 插件。

## 市场结构

```
├── .claude-plugin/plugin.json    # 市场根级插件清单
├── plugins/                      # 独立插件目录
│   └── <plugin-name>/
│       ├── .claude-plugin/plugin.json
│       ├── commands/             # 命令组件
│       ├── agents/               # 子代理组件
│       ├── skills/               # 技能组件
│       │   └── <skill-name>/SKILL.md
│       ├── hooks/                # 钩子组件
│       │   ├── hooks.json
│       │   └── scripts/
│       └── .mcp.json             # MCP 服务配置（可选）
```

## 添加新插件

- 复制 `plugins/plugin-template` 作为起点
- 修改 `.claude-plugin/plugin.json` 中的 name/description
- 按需创建组件目录，不要为了"看起来完整"创建空目录
- 目录和文件名统一使用 `kebab-case`

## 关键路径

- 插件内部引用使用 `${CLAUDE_PLUGIN_ROOT}` 而非硬编码绝对路径
- 全局配置路径：`~/.claude/settings.json`

## 已注册插件速览

| 插件 | 主要组件 |
|------|----------|
| `plugins/apifox-codegen-plugin/` | skills: `apifox-dev`, `generate-interfaces-from-code`, `generate-scenario-tests` |
| `plugins/backend-construct-plugin/` | agents: `backend-plan-agent`; skills: `backend-dev`, `write-plans-with-construct` |
| `plugins/betterpowers/` | hooks; skills: `brainstorming`, `test-driven-development`, `systematic-debugging`, `writing-plans`, `requesting-code-review`, `subagent-driven-development` 等 |

`betterpowers` 有独立的 `hooks/hooks.json`，可通过 SessionStart hook 自动加载。
