# Task/Step 粒度 + 防占位符纪律设计（两条 plan 链路）

> 日期：2026-06-13
> 范围：betterpowers（`brainstorming`、`writing-plans`）+ backend-construct-plugin（`write-plans-with-construct`、`backend-plan-agent`）

## 1. 问题

本设计解决两条 plan 链路上的两类纪律缺口：

**(a) 任务粒度** 目前在 plan 阶段才由 writing-plans 现场从架构单元推导，而该步骤比 brainstorming 拥有更少设计上下文，导致粒度易失真：

- 任务过大 → 一个任务塞进多个行为，reviewer 无法对单一需求判定，实现者上下文扛不住（尤其破坏混合 SDD 的"按任务推进/按任务 fresh reviewer"）。
- 任务过小 → 碎片化，提交/审查 churn。
- 单元过大才在 plan 阶段暴露（本该在 spec 阶段就发现的范围问题）。

**(b) 防占位符** `writing-plans` 已有强力的 No Placeholders 段，但 `write-plans-with-construct` **没有**专门的防占位符纪律，后端 plan 容易出现"参考 example 自行实现"这类空泛任务。

**(c) plan 可评审性** plan 为"无上下文实现者可执行"而写满代码，结果**人难以仔细 review**；而 spec 可评审。这是受众错配——plan 是机器执行件，不是人类评审件。解法不是砍代码（会破坏无上下文可执行性），而是给 plan 一层**可略读的评审面**：让人在"形状/覆盖"高度评审，代码正确性交给 agent（`plan-doc reviewer`）+ SDD 执行期逐任务 review。

目标：让**两条 plan 链路在"任务粒度""防占位符""plan 可评审性"上都有良好表现**，同时保持角色分工——`writing-plans` 泛用、`write-plans-with-construct` 后端专用。

## 2. 已确认决策

| 维度 | 决策 |
|------|------|
| spec 阶段"任务分解概览" | **按复杂度分级**：简单任务一句话；复杂任务才做粗分解+依赖顺序+sizing 检查 |
| 粒度纪律落地形式 | **强制 plan 结构 + skill 指引**：plan 必须含任务清单，每个任务标注粒度依据；同时在 skill 指引写明 task/step 标准 |
| 后端链路改动范围 | **`write-plans-with-construct` + `backend-plan-agent`** 都改 |
| 防占位符 | 两条链路都要有；betterpowers 已有为权威，backend 补齐对齐版本 |
| plan 可评审性 | 每个 task 在代码块前置可略读**评审头**（Intent / Covers spec / Files / Granularity rationale）；人评审 plan 形状，代码块归为执行细节 |
| 角色分工 | 保持：betterpowers 泛用、backend 专用；各项纪律标准在 betterpowers 权威化，backend 自包含对齐复述 |

## 3. 通用纪律：task/step 粒度 + 防占位符（权威来源置于 betterpowers）

这两套标准是两条链路共享的"度量衡"，在 `writing-plans`（及 `brainstorming` 的分解概览）中权威表述；backend 两个文件按领域复述对齐。

### 3.1 Task / Step 粒度

#### Task（任务）
一个**连贯、可独立提交**的单元，驱动**一个可验证行为/结果**，触及一组有界文件，小到实现者能一次性纳入上下文、reviewer 能对着单一需求判定。

- **过大信号**：含一个以上彼此独立的行为；触及多组不相关文件；无法用一个聚焦的失败测试表达；审查要跨多个关注点。→ 拆分。
- **过小信号**：一个 step 被当成 task（如"重命名变量"独立成任务）；制造 churn。→ 合并。
- **经验法则**：一个 task ≈ 一个由测试驱动的行为切片。

#### Step（步骤）
task 内部的一个 2–5 分钟单一动作（写失败测试 / 运行看它失败 / 最小实现 / 运行通过 / 提交）。`writing-plans` 现有的 bite-sized step 结构已达标，保留不动。

### 3.2 防占位符纪律

plan 的每个 step 必须给出工程师真正需要的具体内容。以下为"计划失败"写法，两条链路一律禁止：

- `TBD` / `TODO` / "之后实现" / "待补充" / "fill in details"
- "加上适当的错误处理 / 校验 / 边界情况"这类空泛指令
- "为上面写测试"却不给实际测试代码
- "Similar to Task N"（应重复代码——读者可能乱序阅读）
- 只说做什么、不说怎么做（代码类 step 必须含代码块）
- 引用任何未在任一 task 中定义的类型 / 函数 / 方法

`writing-plans` 已有此段，保留为**权威**；backend 链路需补齐**领域化的对齐版本**（见 §6、§7）。

### 3.3 任务评审头（可略读评审面）

每个 task 在其代码块**之前**必须有一个 2–4 行评审头，让人能在"形状/覆盖"高度评审 plan，而不必逐行读代码：

- **Intent**：这个 task 达成的那一个可验证行为
- **Covers spec**：对应 spec 的哪一节/哪条需求（形成 spec↔plan 覆盖可追溯）
- **Files**：动到的文件
- **Granularity rationale**：为什么这是一个 task，而非多个/半个

评审分工：**人**读评审头，审"覆盖性 / 排序 / 粒度是否失真"；**代码块**由 `plan-doc reviewer`（agent）+ SDD 执行期逐任务 review 把关，不靠人逐行读。

## 4. spec 阶段：Task Decomposition Overview（分级）

在 `brainstorming` 产出的 spec 中新增一节 **Task Decomposition Overview**，作为 spec→plan 的桥，让下游 plan skill **消费**它而非从零重推。

- **简单**（单一连贯单元、路径清晰、任务数少）：一句话——"单一可成 plan 的单元；由 plan skill 切分为任务。"
- **复杂**（跨多个单元 / 拆法不唯一 / 顺序影响返工）：给出
  1. 粗粒度工作分解：行为级任务列表（**不写 bite-sized step**）
  2. 依赖顺序：先做什么能降低返工/不确定性
  3. **sizing 检查**：每项能否在一个 plan 内完成；任何一项过大 → 拆成 sub-project/子 spec

**复杂判定信号**（复用现有概念）：涉及多个组件/单元、拆法不唯一、跨切面依赖，或 "Design for isolation" 已产出多于约定数量的单元。

> spec **不放** bite-sized step——那是 plan 阶段的执行机制。spec 只锁"行为级粗拆分+顺序+sizing"。

## 5. `writing-plans`（泛用）改动

- **强制任务清单 + 评审头**：plan 必须显式呈现任务清单；每个 Task 在代码块前置 §3.3 评审头（Intent / Covers spec / Files / Granularity rationale）。
- **task-sizing 标准**：把 §3.1 的过大/过小信号写入指引，指导拆分判断。
- **防占位符**：`writing-plans` 已有 No Placeholders 段，保留为权威，不削弱。
- **消费 spec 的 Task Decomposition Overview**：若 spec 标了复杂并给了分解，按其行为级列表与顺序细化为 task+step；若标了简单，则按单元切分；若 spec 缺该节（旧 spec），自行推导但套用同一套 §3 标准。
- **保留**：bite-sized step 结构、No Placeholders、Self-Review、plan-doc reviewer。
- **Self-Review 新增一条**：每个 task 恰好对应一个可验证行为；无 task 捆绑多行为；无 step 被提升成 task。

## 6. `write-plans-with-construct`（后端专用）改动

- **层内 task-sizing 标准**（自包含复述 §3.1，对齐措辞）：一个 task 是命中层内的**一个可验证行为**（或行为天然跨层时的一条薄的跨层切片），不是"把整层 X 的活打成一坨"。
- **强制粒度依据 + 评审头**：现有"分层实施任务"中，每层任务块按行为列出多个 task，每个 task 在代码块前置 §3.3 评审头（领域化：`Covers spec` 用对应层需求，`Files` 用实际文件），让后端 plan 同样能在形状高度被人评审。
- **新增防占位符纪律**（领域化复述 §3.2）：每个层任务必须给出具体变更内容——实际函数/方法签名、结构体字段、路由注册写法、实际文件路径与测试——而不是"参考 example 自行实现 service 逻辑"这类空泛任务。`examples/` 是模板，必须为本任务**实例化**，不能把"见 examples/"当作整个任务内容。
- **保留**：层裁剪（涉及层/不涉及层）、E2E-first 验证、先搜现有测试 `Init()` 等既有规则不变。

## 7. `backend-plan-agent`（后端复杂任务拆分）改动

复杂后端任务的实际拆分发生在这里，必须先产出良好粒度，才能让 `write-plans-with-construct` 细化时不失真。

- 其 **"推荐任务拆分"** 输出对齐 §3.1：每个子任务一个可验证行为、文件有界、按依赖排序。
- Working Rules / Output 加入 §3.1 的 sizing 检查（过大拆、过小并）。
- **防占位符（§3.2 的拆分层适配）**：每个子任务必须**具体可落地**——点明行为、命中层、涉及文件/边界，不得用"处理剩余逻辑""其余按需补齐"这类空泛子任务交回下游。
- **对齐评审头字段**：子任务输出已含"行为 / 标签 / 文件"，天然对应 §3.3 的 Intent / Covers / Files；交回时保留这些字段，便于 `write-plans-with-construct` 直接生成评审头。
- 保持其既有边界：不扩标签、不产实现代码、交回 `write-plans-with-construct`。

## 8. 跨插件一致性与漂移风险

两类纪律（粒度 + 防占位符）的**权威文本在 betterpowers**（`writing-plans` + `brainstorming`）。backend 两文件是**对齐的自包含复述**——因为跨插件无法在运行时可靠引用对方文件，skill 必须自包含。

> **漂移风险（须显式记录）：** 日后修改 §3.1 粒度标准或 §3.2 防占位符标准，必须**同步**更新 backend 两文件的复述；二者无法自动跟随。

## 9. 不做的事（YAGNI）

- 不新增独立"任务分解"skill/step（分级入 spec + 复杂走现有 `backend-plan-agent` 已足够）。
- 不改 `subagent-driven-development` / `executing-plans`（它们消费 task，更好的粒度自然流过）。
- 不动 backend 的 knowledge/references/examples 领域三级资产。
- 不把后端层标签等领域内容塞进 betterpowers core。
- 无第三方依赖。

## 10. 文件改动清单

1. `betterpowers/skills/brainstorming/SKILL.md` — 新增 Task Decomposition Overview（分级）+ checklist 项。
2. `betterpowers/skills/writing-plans/SKILL.md` — 强制任务清单+粒度依据、task-sizing 标准、消费 spec 概览、Self-Review 新增项。
3. `backend-construct-plugin/skills/write-plans-with-construct/SKILL.md` — 层内 task-sizing 标准 + 每层任务块强制粒度依据。
4. `backend-construct-plugin/agents/backend-plan-agent.md` — 拆分输出对齐 sizing 标准。

## 11. 验证方式

遵循"零 token 成本本地测试"约束，不新增调真实模型的行为测试：

- 文本一致性 grep：§3.1 粒度、§3.2 防占位符、§3.3 评审头三套标准在 4 个文件中表述对齐、无矛盾；betterpowers 为权威、backend 为复述且语义一致；确认 `write-plans-with-construct` 与 `backend-plan-agent` 确有防占位符段落，且两个 plan skill 都要求每 task 含评审头（Intent / Covers spec / Files / Granularity rationale）。
- 人工走查两个样例：①简单单元任务走 betterpowers 链（spec 一句话概览 → plan 任务清单粒度合理、含评审头、无占位符，且只读评审头即可判断覆盖性）；②复杂后端任务走 backend 链（backend-plan-agent 拆分粒度+具体 → write-plans-with-construct 细化、每层任务含评审头、无"见 examples 自行实现"空泛任务）。
- 确认未触碰 §9 列出的不改项；无第三方依赖。
