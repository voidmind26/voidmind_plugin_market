---
name: backend-dev
description: 用于后端开发计划增强。触发于“新增后端接口”“修改 service / data / controller / router”“只改某几层后端逻辑”“需要先判断后端变更范围再写 plan”等场景。负责识别后端任务、推断并确认分层标签，并将结果路由到后续的 plan skill 或复杂任务辅助 agent。
argument-hint: [需求描述或变更范围]
allowed-tools:
  - Read
  - Edit
  - Write
  - Bash
  - AskUserQuestion
  - Skill
  - Agent
---

# Backend Dev

将后端开发需求先做分层识别，再把已确认的分层约束交给后续计划生成流程。该 skill 是统一入口，不直接替代具体实现 skill，也不直接生成完整 plan。

## Goals

- 识别当前任务是否属于后端开发。
- 在固定原子标签集合内推断最小必要变更面。
- 区分“已确认层”“候选但未确认层”“明确不涉及层”。
- 根据复杂度决定：直接交给 `write-plans-with-construct`，或转交 `backend-plan-agent`。
- 输出稳定的结构化结果，供后续 plan skill 或 plan agent 消费。

## Atomic Labels

仅使用以下固定原子标签，不扩展同义新标签：

- `dto`
- `data`
- `service`
- `controller`
- `router`
- `task`

标签含义保持稳定：

- `dto`：请求/响应结构、参数绑定字段、返回字段定义。
- `data`：MySQL/Redis/DAO/Repository 读写与查询封装；若需求涉及 model、持久化实体、表结构映射或缓存对象结构，也统一并入 `data`。
- `service`：业务编排、业务规则、跨数据源调用。
- `controller`：HTTP 参数绑定、调用 service、返回响应。
- `router`：路由注册、方法与路径挂载、中间件接线。
- `task`：定时任务、周期任务、任务入口与调度接线。

## Required Behavior

1. 先判断任务是不是后端开发。
   - 属于后端开发：继续做分层识别。
   - 明显不是后端开发：立即退出，并说明该 skill 不接管，例如纯前端页面、纯文档、纯发布配置、纯测试用例整理。

2. 只按最小变更面推断标签。
   - 默认从用户明确提到的层开始推断。
   - 不因“常见链路”自动扩成完整 `dto + data + service + controller + router`。
   - 只有当需求本身明确要求新增接口、补齐接线、或用户确认需要全链路时，才扩大标签集合。

3. 先确认，再写 plan。
   - 未确认层不得自动写入 plan。
   - 对措辞里带“可能”“顺便”“看情况”的层，先标成候选，不直接并入 `confirmed_labels`。
   - 无法唯一判断时，先追问或升级，不抢写结论。

4. 明确简单与复杂分流，并由当前 skill 主动继续路由。
   - 若标签组合跨 3 层以上，转交 `backend-plan-agent`。
   - 若需求模糊、拆法不唯一、边界依赖不清，也转交 `backend-plan-agent`。
   - 若任务简单，直接把已确认标签、涉及层、不涉及层、适用规范来源交给 `write-plans-with-construct`。
   - 不要停在“给出判断结果”等外层继续决定；`backend-dev` 自己负责在产出固定结构后，继续调用下游 skill 或 agent。

5. 输出必须体现边界。
   - 给出已确认标签，并以它作为后续 plan 的唯一权威标签集合。
   - 给出用户明确提及的层、候选但未确认的层、以及不涉及层。
   - 对未纳入的层说明原因：未提及、用户限定不改、当前信息不足、与最小变更面不符。

6. 规范来源不要在本文件内手写大段映射，统一遵循：
   - 先服从 `knowledge/layering.md` 的全局原则。
   - 再按 `references/knowledge-map.md` 为命中标签选择 `knowledge/`、`references/`、`examples/`。
   - 只把命中标签真正需要的路径加入 `knowledge_sources`，不要把全部资料一股脑塞给下游。

## Output Contract

在继续调用下游之前，先产出以下唯一 JSON 形态的固定结构，作为当前 skill 向下游 skill / agent 传递的内部载荷；字段名与字段顺序保持不变：

```json
{
  "status": "backend | declined",
  "confirmed_labels": [],
  "involved_layers": {
    "confirmed": [],
    "candidate": []
  },
  "excluded_layers": [
    {
      "layer": "dto",
      "reason": "not_mentioned"
    }
  ],
  "task_summary": "",
  "complexity": "simple | complex | declined",
  "knowledge_sources": [],
  "next_skill": "write-plans-with-construct | backend-plan-agent | none"
}
```

固定结构说明：

- `status`
  - `backend`：当前任务属于后端 plan 范围，由本 skill 接管并继续路由。
  - `declined`：当前任务明显不属于后端 plan 范围；此时不再做分层判断。
- `confirmed_labels`: 原子标签数组，表示最终确认纳入计划的层，也是后续 plan 生成唯一应消费的权威标签集合。
- `involved_layers`
  - `confirmed`: 原子标签数组，必须与 `confirmed_labels` 完全一致，仅用于给下游携带分组后的结构，不得与 `confirmed_labels` 产生差异
  - `candidate`: 原子标签数组，表示可能涉及但未确认的层
- `excluded_layers`: 对象数组，每项结构为
  - `layer`: 原子标签
  - `reason`: `not_mentioned | user_excluded | insufficient_context | violates_minimum_scope`
- `task_summary`: 字符串，概括本次已确认的任务意图，供 `write-plans-with-construct` 直接消费。
- `complexity`
  - `simple`：边界清楚，标签不超过 3 层，适合直接交给 `write-plans-with-construct`
  - `complex`：跨 3 层以上，或拆分路径不唯一，或存在明显边界不确定性，应先转交 `backend-plan-agent`
  - `declined`：非后端任务，不进入后续后端 plan 路由
- `knowledge_sources`: 字符串数组，只使用插件内知识资产路径，来源由 `references/knowledge-map.md` 决定，允许来自 `knowledge/*.md`、`references/*.md`、`examples/*.md`，但必须与命中层直接相关。
- `next_skill`
  - `write-plans-with-construct`：当 `status = backend` 且 `complexity = simple`
  - `backend-plan-agent`：当 `status = backend` 且 `complexity = complex`；它是中间辅助步骤，不是终点，输出应回到 `write-plans-with-construct`
  - `none`：当 `status = declined`

各状态的空值规则也必须固定：

- `status = declined`
  - `confirmed_labels = []`
  - `involved_layers.confirmed = []`
  - `involved_layers.candidate = []`
  - `excluded_layers = []`
  - `task_summary = ""`
  - `knowledge_sources = []`
  - `complexity = declined`
  - `next_skill = none`
- `status = backend` 且 `complexity = simple`
  - 必须至少给出一个 `confirmed_labels`
  - `task_summary` 必须是可直接交给 `write-plans-with-construct` 的一句话任务摘要
  - `next_skill = write-plans-with-construct`
- `status = backend` 且 `complexity = complex`
  - `confirmed_labels` 可为空或非空，但必须解释候选层与排除层
  - `task_summary` 也必须是可直接交给下游的一句话任务摘要
  - `next_skill = backend-plan-agent`

产出该结构后，当前 skill 继续主动调用 `next_skill`，而不是把路由责任留给外层猜测。

## Examples

### 例 1：只改查询逻辑

输入意图：只改查询逻辑，不动接口出参。

候选标签：`service + data`

判断要点：
- 查询条件拼装、数据读取通常落在 `data`。
- 查询结果整理、业务过滤通常落在 `service`。
- 未明确提到返回结构变化时，不自动带上 `dto` 或 `controller`。

### 例 2：只改返回字段和参数绑定

输入意图：只改返回字段和参数绑定，不动业务逻辑。

候选标签：`dto + controller`

判断要点：
- 参数绑定与响应拼装在 `controller`。
- 请求/响应字段定义在 `dto`。
- 未提到规则变化时，不自动带上 `service`。

### 例 3：新增接口

输入意图：新增一个后端接口。

候选标签：`dto + data + service + controller + router`

判断要点：
- 新增接口通常需要新增请求响应定义、业务逻辑、数据访问、控制器接线、路由注册。
- 这是典型多层任务；若边界清楚但已跨 3 层以上，仍按复杂任务转交 `backend-plan-agent`。

### 例 4：补定时任务逻辑

输入意图：补一个定时任务逻辑。

候选标签：`task + service`

判断要点：
- 调度入口、任务注册落在 `task`。
- 任务真正执行业务通常落在 `service`。
- 若只是补任务接线，不自动带上 `data`；除非需求明确涉及数据读写改造。
