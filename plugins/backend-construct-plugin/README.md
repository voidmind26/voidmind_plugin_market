# backend-construct-plugin

面向后端 plan 阶段的 Claude Code 插件骨架，用于在开始实现前先收敛计划范围、确认涉及的后端层，并选择合适的规划入口。

## 边界

- 仅覆盖后端场景
- 仅介入 plan 阶段，不覆盖编码、联调、提测等后续流程
- agent 仅在复杂任务时作为辅助，不是默认路径

## 核心概念

- **分层标签**：由固定原子层和灵活组合组成。原子层覆盖 dto、data、service、controller、router、task 等后端层；若需求涉及持久化实体、表结构映射或缓存对象结构，统一并入 `data` 处理。组合仅用于确认本次 plan 实际涉及哪些层。
- **`backend-dev`**：统一入口 skill，负责进入后端 plan 工作流、做初步标签判断与路由，不直接承担完整 plan 生成。
- **`write-plans-with-construct`**：标签驱动的 plan skill，根据分层标签产出实现计划。
- **`backend-plan-agent`**：复杂任务辅助 plan agent，在任务跨度大、依赖多或需要拆解阶段时介入。

## Workflow

1. 先通过 `backend-dev` 进入后端 plan 工作流。
2. 用分层标签确认本次计划涉及的后端层。
3. 常规任务直接进入 `write-plans-with-construct` 生成计划。
4. 复杂任务再引入 `backend-plan-agent` 辅助拆解，然后回到 `write-plans-with-construct` 落计划。

## 提供什么

- 后端 plan 阶段统一入口
- 基于分层标签的计划生成
- 复杂任务场景下的辅助规划能力
- `references/knowledge-map.md` 作为统一索引，维护标签到 `knowledge/`、`references/`、`examples/` 的路由规则
- `examples/` 下的最小代码模板，用于承接不适合放进主 knowledge 的具体写法示例，包括 DTO、Model/Data、Controller/Router、Task/Logging 等场景

## 知识结构

- `knowledge/layering.md`：全局原则文档，定义最小变更面与分层职责边界。
- `references/knowledge-map.md`：统一索引文档，定义标签如何路由到 `knowledge/`、`references/`、`examples/`。
- `knowledge/`：高层规则与边界约束。
- `references/`：细规则、目录约定、禁忌项、签名模式。
- `examples/`：最小代码模板，用于需要具体写法时提供参考。
