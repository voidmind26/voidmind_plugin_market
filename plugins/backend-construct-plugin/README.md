# backend-construct-plugin

面向 Go 后端代码开发的独立规范插件。它从目标仓库提取真实约定，按最小分层范围加载规范，直接生成或修改代码并完成验证。

## 边界

- 覆盖 DTO、model/data、service、controller、router、task 和日志相关的后端开发。
- 不绑定任何计划目录、执行模式、subagent 或外部工作流插件。
- 不强推固定框架；目标项目的仓库指令和现有实现优先于插件示例。
- 不因常见完整链路自动扩大修改范围。

## 核心技能

- **`backend-dev`**：识别必要分层，读取项目代码与对应规范，直接实施代码修改并验证。

## 工作方式

1. 读取仓库指令、工作区状态、目标调用链和相邻实现。
2. 用 `dto / data / service / controller / router / task` 标签确定最小变更面。
3. 通过 `references/knowledge-map.md` 渐进加载命中层资料。
4. 按项目现有结构直接实现代码，同步必要调用方与测试。
5. 从聚焦检查到真实业务链路完成验证。

## 知识结构

- `knowledge/layering.md`：分层职责和最小变更面。
- `references/knowledge-map.md`：标签到规范资产的读取索引。
- `knowledge/`：各层高层规则与边界约束。
- `references/`：字段、签名、目录、禁忌项和验证准则。
- `examples/`：项目缺少相邻实现时使用的最小代码形态参考。

示例不是模板真相。若示例与目标项目冲突，以用户要求、仓库指令和项目现有代码为准。
