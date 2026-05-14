# voidmind_plugin_market

私有 zcode 插件集合仓库。

## 目录结构

```
├── .claude-plugin/
│   └── plugin.json          # 市场级插件清单
├── plugins/                 # 插件目录，每个子目录为一个独立插件
│   ├── apifox-codegen-plugin/     # Apifox 接口文档生成
│   ├── backend-construct-plugin/  # 后端 plan 统一入口
│   ├── betterpowers/              # 优化版 superpowers
│   └── plugin-template/           # 插件模板（新建插件时复制）
├── CLAUDE.md
└── README.md
```

## 已注册插件

| 插件 | 版本 | 说明 |
|------|------|------|
| [apifox-codegen-plugin](plugins/apifox-codegen-plugin/) | 0.1.0 | 从 Go Web 代码生成 Apifox 接口文档与场景化测试用例 |
| [backend-construct-plugin](plugins/backend-construct-plugin/) | 0.2.0 | 后端 plan 阶段统一入口，提供标签驱动计划与复杂任务辅助 |
| [betterpowers](plugins/betterpowers/) | 1.0.0 | 优化版 superpowers，包含多项开发技能与工作流 |

## 添加新插件

1. 复制 `plugins/plugin-template` 到 `plugins/<plugin-name>/`
2. 修改 `plugins/<plugin-name>/.claude-plugin/plugin.json`
3. 按需创建 `commands/`、`agents/`、`skills/`、`hooks/` 等组件目录
4. 更新本 `README.md` 注册插件

## 插件规范

每个插件需包含 `.claude-plugin/plugin.json`，遵循标准 zcode 插件结构。
