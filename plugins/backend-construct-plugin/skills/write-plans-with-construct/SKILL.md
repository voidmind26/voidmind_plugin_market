---
name: write-plans-with-construct
description: 用于后端开发计划生成。输入是已确认的分层标签与后端规范来源，输出只覆盖命中层的实现 plan，并显式声明不涉及层。触发于 backend-dev 已完成标签确认之后。
argument-hint: [confirmed_labels + task summary]
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
---

# Write Plans With Construct

将已确认的后端标签组合写成可执行计划。该 skill 只负责按标签裁剪 plan，不默认展开完整链路。

## Required Inputs

输入至少要包含：

- `confirmed_labels`
- `task_summary`
- `involved_layers`
- `excluded_layers`
- `knowledge_sources`

输入约束：

- `confirmed_labels` 是计划裁剪的唯一权威标签集合。
- `involved_layers.confirmed` 若存在，必须与 `confirmed_labels` 一致。
- `excluded_layers` 只能列出明确不纳入本次 plan 的层，不允许再为这些层创建默认任务。
- `knowledge_sources` 只列出与命中层有关的插件内知识资产路径，可来自 `knowledge/*.md`、`references/*.md`、`examples/*.md`，具体选择规则由 `references/knowledge-map.md` 统一定义。
- 生成 plan 前，先读 `knowledge/layering.md`；再按 `references/knowledge-map.md` 为当前命中标签选择真正需要的 `knowledge/`、`references/`、`examples/`。

## Hard Rules

- 如果标签未覆盖某层，plan 不得为该层创建默认任务。
- plan 必须显式列出涉及层与不涉及层。
- 验证步骤按命中层裁剪。
- 若计划涉及测试、补测、E2E、集成验证或单测实现，先要求模型在代码中搜索现有测试初始化方式；默认优先复用框架测试包下的 `Init()` 入口。只有在确认仓库不存在统一 `Init()` 时，才退回到项目现有测试样式补齐初始化，不直接规划手写初始化或自行拼装测试环境。
- 验证步骤默认先写命中层的最高自然层级验证；存在真实业务链路时，优先写 E2E 或集成验证。
- 只有在大粒度验证失败后需要诊断、当前命中层天然没有自然 E2E 路径，或存在更自然的集成边界时，才补充更小粒度测试。
- 若补充了小粒度测试，最终仍要回到大粒度验证闭环。
- 只有接近完整接口新增时，才参考 `construct-framework` 的全链路顺序。
- `excluded_layers` 只用于声明范围边界，不用于补充任务。
- 若 `confirmed_labels` 只命中单层或局部组合，必须保持局部视角，不得回退到默认完整链路。

## Plan Sections

生成的 plan 至少包含以下部分：

1. 任务摘要
2. 已确认标签
3. 涉及层 / 不涉及层
4. 分层实施任务
5. 适用规范来源
6. 验证步骤

要求：

- “涉及层 / 不涉及层”必须直接使用上游输入，不重新扩面。
- “分层实施任务”只为 `confirmed_labels` 命中的层创建任务块。
- “适用规范来源”只引用本次真实使用的知识来源。
- 若命中层只需要边界规则，读 `knowledge/` 即可；若需要细规则，继续读 `references/`；若需要具体写法模板，再读 `examples/`。
- “验证步骤”只覆盖命中层会触发的检查，不补充未命中层的默认验证。
- “验证步骤”若涉及测试实现，必须先写明：先搜索项目内测试初始化方式，默认优先复用框架测试包下的 `Init()`；仅在确认不存在统一 `Init()` 时，才遵循仓库现有测试样式补齐初始化。
- “验证步骤”默认从命中层的最高自然层级开始：有自然业务链路时先写 E2E/集成验证，不先从单元测试起手。
- 若需要更小粒度测试，必须把用途限定为诊断失败原因，或覆盖当前命中层天然没有端到端路径的局部行为。
- 若补充了小粒度测试，最终仍要回到大粒度验证闭环。

## Layer Behavior

按命中层拆 plan：

- `service`
  - 写业务编排、错误处理、日志与依赖调用任务。
  - 若未命中 `data`，不要补数据访问改造任务。
- `data`
  - 写查询、存储、缓存访问边界任务。
  - 若需求涉及 model、持久化实体或缓存对象结构，也统一按 `data` 处理。
  - 若未命中 `service`，不要默认补业务逻辑编排。
- `dto`
  - 写输入输出结构任务。
  - 若未命中 `controller`，不要自动补参数绑定调整。
- `controller`
  - 写参数绑定与响应任务。
  - 若未命中 `router`，不要默认补路由注册。
- `router`
  - 写路由注册任务。
  - 若未命中 `controller`，不要自动补 handler 重构。
- `task`
  - 写任务入口与调度任务。
  - 若未命中 `service`，不要默认补任务内业务逻辑改造。

## Examples

### Example: `service+data`

- `confirmed_labels`: `service+data`
- 涉及层：`service`、`data`
- 不涉及层：`dto`、`controller`、`router`、`task`
- `knowledge_sources`: `knowledge/service.md`, `knowledge/data.md`, `references/service-conventions.md`, `references/model-conventions.md`, `examples/model-data-example.md`

计划约束：

- 只创建 `service` 与 `data` 的任务块。
- 不得默认创建 `controller/router` 任务。
- 验证步骤先检查业务逻辑与数据访问链路的集成验证，不补 HTTP 接入层检查。
- 只有当该集成验证暴露定位困难时，才补 `service` 或 `data` 侧的小粒度测试。
- 即使补了小粒度测试，最终仍要回到业务链路级验证闭环。

### Example: `controller+router`

- `confirmed_labels`: `controller+router`
- 涉及层：`controller`、`router`
- 不涉及层：`dto`、`service`、`data`、`task`
- `knowledge_sources`: `knowledge/controller-router.md`, `references/controller-router-conventions.md`, `examples/controller-router-example.md`

计划约束：

- 只创建 `controller` 与 `router` 的任务块。
- 不得默认追加业务逻辑重构。
- 验证步骤优先覆盖参数绑定、响应与路由注册的接口级 E2E/集成验证。
- 只有当接口级验证失败且定位不清时，才补 controller 或 router 侧的小粒度测试。
- 即使补了小粒度测试，最终仍要回到接口级验证闭环。

### Example: `dto+service+controller+router`

- `confirmed_labels`: `dto+service+controller+router`
- 涉及层：`dto`、`service`、`controller`、`router`
- 不涉及层：`data`、`task`

计划约束：

- 仍然只为命中的 4 层创建任务。
- 即便接近完整接口，也不能自动补 `data`，除非标签已确认。
- 验证步骤优先覆盖该 4 层共同组成的接口级 E2E/集成验证。
- 若接口级验证失败后仍难以定位，再按实际失败信号下钻到对应层的小粒度测试。
- 只有在任务接近完整接口新增时，才参考插件内整理后的完整链路顺序；否则始终按命中层拆解。

## Pressure Tests

以下场景用于校验 plan 是否默认先产出最高自然层级验证，而不是回退到单测优先。

### Baseline Failure 1

输入任务：新增一个查询接口，确认标签 `service + data + controller + router`。

错误产物特征：
- 验证步骤先列“为 service 写单测”“为 data 写单测”
- 没有先给接口请求到响应的 E2E/集成验证
- 把 E2E 放到最后作为“补充”

### Baseline Failure 2

输入任务：补一个定时任务入口，确认标签 `task + service`。

错误产物特征：
- 直接拆成多个函数级单测
- 没有先给任务入口触发到业务执行的集成验证
- 没有要求在诊断后回到任务链路验证

### Expected Pass Signal

正确产物必须满足：
- 先写命中层的最高自然层级验证
- 有自然业务链路时，默认先写 E2E 或集成验证
- 只有当大粒度验证失败、定位不清或天然无 E2E 路径时，才补小粒度测试
- 补了小粒度测试后，最终仍回到大粒度验证闭环

### Scenario 1: HTTP query flow

输入任务：新增查询接口，确认标签 `service + data + controller + router`

通过标准：
- 计划先给接口请求到响应的 E2E/集成验证
- 计划不会默认先列 `service` / `data` 单测
- 若补更小粒度测试，必须明确是因为接口级验证失败后需要进一步定位
- 最终仍要回到接口级验证闭环

### Scenario 2: Scheduled task flow

输入任务：新增定时任务入口，确认标签 `task + service`

通过标准：
- 计划先给任务入口触发到业务执行的集成验证
- 若补小粒度测试，必须显式说明是为定位失败原因
- 最终仍要回到任务链路验证

### Scenario 3: Natural no-E2E path

输入任务：只调整一个纯 `dto` 字段映射规则，确认标签 `dto`

通过标准：
- 计划可以从更小但仍真实的验证边界起步
- 计划不会虚构 HTTP 级 E2E
- 计划会说明当前命中层天然没有自然端到端路径，因此采用局部验证

### Red Flags

- “先把 `service` 单测补齐，再看是否需要集成验证”
- “通常顺便为每层都补一份单测”
- “E2E 最后再补”
- “先写几个小测试更稳妥”
- “虽然有真实链路，但从单测起手更方便”
- “直接手写测试初始化，不先搜索现有 `Init()` 入口”
- “默认自行拼装数据库、缓存、配置等测试环境”

以上均视为失败信号。

## Output Guidance

输出 plan 时，优先使用以下表述：

- “本次确认标签为 ……”
- “本次涉及层为 ……”
- “本次不涉及层为 ……”
- “以下任务仅覆盖已确认标签，不补充未命中层默认任务”

不要输出的内容：

- “通常还应该顺便改 ……”
- “为了完整性一并补上 ……”
- “默认增加 controller/router/data”

## Output & Execution Handoff

将生成的 plan 以 markdown 格式保存到 betterpowers 的 plan 目录：`docs/superpowers/plans/YYYY-MM-DD-<feature-name>.md`。

**重要：** 保存 plan 后，不要进入 Claude Code 内置的 Plan 模式（EnterPlanMode）。后端开发插件的 plan 配合 betterpowers 的 plan 执行链路工作。

保存 plan 后，MUST 询问用户选择执行方式。不要自动选择——即使用户明确提到"执行"或"实现"，也要先让用户选择。

**"Plan 已保存到 `docs/superpowers/plans/<filename>.md`。**

**此计划包含 [N] 个任务，涉及 [labels]。选择执行方式：**

**1. Subagent-Driven（推荐）** — 使用固定角色 subagent 执行，包含 spec review 和 code quality review

**2. Inline 执行** — 在当前会话中逐任务执行，分批推进并在检查点确认

**选择哪种方式？"**

等待用户明确选择后再继续。不要假设或默认。

**若选择 Subagent-Driven：**
- **REQUIRED SUB-SKILL:** Use superpowers:subagent-driven-development
- 固定角色 subagent + 两阶段 review

**若选择 Inline 执行：**
- **REQUIRED SUB-SKILL:** Use superpowers:executing-plans
- 分批执行并在检查点确认
