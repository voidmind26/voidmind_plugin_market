# 后端开发插件测试初始化提示增强实现计划

## 1. 任务摘要

修改 `write-plans-with-construct` skill 的规则文案，在 plan 中明确要求模型先搜索现有测试初始化方式，默认优先使用框架测试包下的 `Init()` 方法，避免因错误初始化反复循环。

## 2. 已确认标签

本次确认标签为：`service`

## 3. 涉及层 / 不涉及层

本次涉及层为：`service`

本次不涉及层为：

- `dto`（not_mentioned）
- `data`（not_mentioned）
- `controller`（not_mentioned）
- `router`（not_mentioned）
- `task`（not_mentioned）

以下任务仅覆盖已确认标签，不补充未命中层默认任务。

## 4. 分层实施任务

### service

1. 定位 `skills/write-plans-with-construct/SKILL.md` 中与验证步骤相关的约束段落，优先检查 `Hard Rules`、`Plan Sections`、`Pressure Tests` 与 `Red Flags`。
2. 在 `Hard Rules` 中补充测试初始化规则，明确：
   - 当计划涉及测试、补测、E2E、集成验证或单测实现时，先在代码中搜索现有测试初始化方式；
   - 默认优先复用框架测试包下的 `Init()` 入口；
   - 只有在确认仓库不存在统一 `Init()` 时，才退回到项目现有测试样式；
   - 不直接规划手写初始化或自行拼装测试环境。
3. 在 `Plan Sections` 的“验证步骤”要求中补充显式输出要求，确保最终生成的 plan 会把这条初始化 tips 写出来，而不是只停留在内部规则。
4. 在 `Pressure Tests` 或 `Red Flags` 中补一条失败信号，约束错误产物不得跳过搜索现有初始化入口，也不得默认要求手写测试初始化。
5. 通读全文，确认新增文案与现有“大粒度验证优先、失败后再下钻”的规则不冲突，也不把 `Init()` 误写成所有仓库都必须存在的强制唯一方案。

## 5. 适用规范来源

- `knowledge/layering.md`
- `knowledge/service.md`
- `references/service-conventions.md`

## 6. 验证步骤

1. 通读修改后的 `skills/write-plans-with-construct/SKILL.md`，确认新增规则只作用于计划中的测试初始化提示，不扩展到未命中层。
2. 检查文案是否明确要求：先搜索项目内测试初始化方式，默认优先复用框架测试包的 `Init()`。
3. 检查文案是否保留回退条件：仅在确认不存在统一 `Init()` 时，才遵循仓库现有测试样式补齐初始化。
4. 检查文案是否避免输出“直接手写初始化”“自行拼装数据库/缓存/配置测试环境”这类错误引导。
5. 结合既有 pressure tests / red flags 复查，确认原有“先最高自然层级验证，再按需下钻”的策略未被破坏。
