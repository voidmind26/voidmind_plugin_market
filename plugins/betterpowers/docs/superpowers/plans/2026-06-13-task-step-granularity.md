# Task/Step 粒度 + 防占位符 + plan 可评审性 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) to implement this plan with fixed-role subagents for implementation, spec review, and code quality review, or superpowers:executing-plans for inline execution. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 betterpowers 与 backend-construct-plugin 两条 plan 链路在"任务粒度""防占位符""plan 可评审性"上都有良好表现。

**Architecture:** 纯技能文档（markdown）改动，跨 2 插件 4 文件。三套纪律（§3.1 粒度 / §3.2 防占位符 / §3.3 评审头）在 betterpowers 权威表述，backend 两文件领域化对齐复述。无生产代码。

**Tech Stack:** Markdown 技能/agent 文件；验证用 `grep` 文本一致性走查。

**说明:** 本 plan 的任务自身采用新设计的"评审头"格式（Intent / Covers spec / Files / Granularity rationale）作为示范。

---

### Task 1: brainstorming — 新增 Task Decomposition Overview（分级）

- **Intent:** 让 spec 阶段产出分级的行为级任务分解概览，作为 spec→plan 的桥
- **Covers spec:** §4
- **Granularity rationale:** 单一文件、单一行为（spec 增加一节指引），独立可验证

**Files:**
- Modify: `plugins/betterpowers/skills/brainstorming/SKILL.md`

- [ ] **Step 1: 在 "Presenting the design" 的 Cover 行追加 Task Decomposition Overview**

将：
```
- Cover: architecture, components, data flow, error handling, testing.
```
替换为：
```
- Cover: architecture, components, data flow, error handling, testing, and a Task Decomposition Overview (see below).
```

- [ ] **Step 2: 在 "**Working in existing codebases:**" 之前插入新小节**

在 `**Working in existing codebases:**` 这一行正上方插入：
```
**Task Decomposition Overview (tiered):**

The spec is the human-reviewable artifact; the plan is not. So decide the coarse task breakdown here, where design context is richest, and let writing-plans refine it into steps. Include a "Task Decomposition Overview" section in the spec, sized to complexity:

- **Simple** (a single coherent unit, clear path, few tasks): one line — "Single plannable unit; the plan skill will slice it into tasks."
- **Complex** (spans multiple units, the split is not unique, or ordering affects rework): provide
  1. a behavior-level task breakdown (NOT bite-sized steps),
  2. the dependency order (what to build first to reduce rework/uncertainty), and
  3. a sizing check — each item must be completable within one plan; if any item is too big, decompose it into a sub-project/sub-spec.

Do NOT put bite-sized implementation steps in the spec — those belong in the plan. The spec only locks the behavior-level breakdown, ordering, and sizing.

```

- [ ] **Step 3: 验证**

Run: `grep -n "Task Decomposition Overview (see below)\|Task Decomposition Overview (tiered)\|a sizing check" plugins/betterpowers/skills/brainstorming/SKILL.md`
Expected: 三处均命中。

---

### Task 2: writing-plans — 新增 Task Sizing and Review Header 小节

- **Intent:** 给 writing-plans 加上任务粒度判据、消费 spec 概览的规则、以及强制评审头
- **Covers spec:** §3.1、§3.3、§5（消费 spec 概览 + sizing + 评审头）
- **Granularity rationale:** 一处新增小节，承载粒度+评审头指引，独立成块

**Files:**
- Modify: `plugins/betterpowers/skills/writing-plans/SKILL.md`

- [ ] **Step 1: 在 "## Bite-Sized Task Granularity" 之前插入新小节**

在 `## Bite-Sized Task Granularity` 这一行正上方插入（注意整段用四反引号围栏，因为内部含三反引号代码块）：

````
## Task Sizing and Review Header

**Consume the spec's Task Decomposition Overview.** If the spec includes one, use its behavior-level breakdown and ordering as the basis for your tasks instead of re-deriving them; refine each item into a task with steps. If the spec marked the work simple (or predates this section), derive the breakdown yourself — but apply the sizing criteria below either way.

**Task sizing.** A task is one coherent, independently committable unit that drives ONE verifiable behavior, touches a bounded set of files, and is small enough to hold in context and to review against a single requirement.

- Too big (split it): more than one independent behavior; touches many unrelated files; can't be expressed by a single focused failing test; review would span multiple concerns.
- Too small (merge it): a single step masquerading as a task (e.g., "rename a variable" as its own task); creates churn.
- Rule of thumb: one task ≈ one test-driven behavior slice.

**Every task leads with a review header**, placed before its steps, so a human can review the plan's shape and coverage without reading every code block:

```
### Task N: <name>
- **Intent:** <the one verifiable behavior this task achieves>
- **Covers spec:** <which spec section/requirement it implements>
- **Files:** <files touched>
- **Granularity rationale:** <why this is one task, not several or half of one>
```

The human reviews these headers (coverage, ordering, sizing); the code in the steps is verified by the plan-document reviewer and by execution-time per-task review.

````

- [ ] **Step 2: 验证**

Run: `grep -n "## Task Sizing and Review Header\|Consume the spec's Task Decomposition Overview\|one test-driven behavior slice\|Every task leads with a review header" plugins/betterpowers/skills/writing-plans/SKILL.md`
Expected: 四处均命中。

---

### Task 3: writing-plans — Task Structure 模板加评审头 + Self-Review 加项

- **Intent:** 让模板默认带评审头、Self-Review 把关粒度与评审头完整性
- **Covers spec:** §5（评审头落到模板）、§5 最后一条（Self-Review 新增项）
- **Granularity rationale:** 与 Task 2 同文件但属不同小节（模板/自检），各自独立可验证

**Files:**
- Modify: `plugins/betterpowers/skills/writing-plans/SKILL.md`

- [ ] **Step 1: 在 Task Structure 模板头部插入评审头三行**

将：
```
### Task N: [Component Name]

**Files:**
- Create: `exact/path/to/file.py`
```
替换为：
```
### Task N: [Component Name]

- **Intent:** [the one verifiable behavior this task achieves]
- **Covers spec:** [which spec section/requirement]
- **Granularity rationale:** [why this is one task, not several or half of one]

**Files:**
- Create: `exact/path/to/file.py`
```

- [ ] **Step 2: Self-Review 增加第 4 条**

将：
```
**3. Type consistency:** Do the types, method signatures, and property names you used in later tasks match what you defined in earlier tasks? A function called `clearLayers()` in Task 3 but `clearFullLayers()` in Task 7 is a bug.

If you find issues, fix them inline. No need to re-review — just fix and move on. If you find a spec requirement with no task, add the task.
```
替换为：
```
**3. Type consistency:** Do the types, method signatures, and property names you used in later tasks match what you defined in earlier tasks? A function called `clearLayers()` in Task 3 but `clearFullLayers()` in Task 7 is a bug.

**4. Task granularity & headers:** Does each task map to exactly one verifiable behavior (no task bundles several; no step promoted to a task)? Does every task have a complete review header (Intent / Covers spec / Files / Granularity rationale)?

If you find issues, fix them inline. No need to re-review — just fix and move on. If you find a spec requirement with no task, add the task.
```

- [ ] **Step 3: 验证**

Run: `grep -n "the one verifiable behavior this task achieves\|Task granularity & headers" plugins/betterpowers/skills/writing-plans/SKILL.md`
Expected: 两处均命中。

- [ ] **Step 4: Commit betterpowers 改动（Task 1-3）**

```bash
git add plugins/betterpowers/skills/brainstorming/SKILL.md plugins/betterpowers/skills/writing-plans/SKILL.md
git commit -m "feat: task sizing, review header, and spec-stage decomposition overview in betterpowers plan flow"
```

---

### Task 4: write-plans-with-construct — 粒度/评审头/防占位符（领域化）

- **Intent:** 让后端写细 plan 的 skill 具备层内任务粒度、评审头、防占位符纪律
- **Covers spec:** §6、§3.3（评审头）、§3.2（防占位符）
- **Granularity rationale:** 单文件，新增一节纪律 + 一处 Plan Sections 引用，是一个连贯行为

**Files:**
- Modify: `plugins/backend-construct-plugin/skills/write-plans-with-construct/SKILL.md`

- [ ] **Step 1: 在 "## Plan Sections" 之前插入新纪律小节**

在 `## Plan Sections` 这一行正上方插入：
```
## 任务粒度、评审头与防占位符

这三条与 betterpowers `writing-plans` 同源（betterpowers 为权威，此处为后端领域化对齐复述；若 betterpowers 标准更新，此处需同步）。

**层内任务粒度。** 一个 task 是命中层内的一个可验证行为（或行为天然跨层时的一条薄跨层切片），不是"把整层 X 的活打成一坨"。
- 过大（拆）：一个 task 含多个独立行为，或无法用一个聚焦验证表达。
- 过小（并）：把一个 step 单独当 task。
- 经验法则：一个 task ≈ 一个由验证驱动的行为切片。

**评审头。** 每层任务块中的每个 task，在代码前置一个可略读评审头，让人不必逐行读代码即可评审 plan 形状：
- **Intent**：该 task 达成的那一个可验证行为
- **Covers**：对应命中层的哪条需求
- **Files**：实际动到的文件
- **Granularity rationale**：为什么这是一个 task

**防占位符。** 每个层任务必须给出具体变更内容——实际函数/方法签名、结构体字段、路由注册写法、实际文件路径与测试，而不是"参考 example 自行实现 service 逻辑"这类空泛任务。`examples/` 是模板，必须为本任务实例化；不得把"见 examples/"当作整个任务内容。禁止 `TBD`/`TODO`/"之后补"/"按需补齐"。

```

- [ ] **Step 2: 更新 Plan Sections 中"分层实施任务"的要求**

将：
```
- “分层实施任务”只为 `confirmed_labels` 命中的层创建任务块。
```
替换为：
```
- “分层实施任务”只为 `confirmed_labels` 命中的层创建任务块；每个 task 遵循“任务粒度、评审头与防占位符”一节（每 task 一个可验证行为 + 评审头 + 具体代码，不留空泛任务）。
```

- [ ] **Step 3: 验证**

Run: `grep -n "## 任务粒度、评审头与防占位符\|一个由验证驱动的行为切片\|遵循“任务粒度、评审头与防占位符”一节" plugins/backend-construct-plugin/skills/write-plans-with-construct/SKILL.md`
Expected: 三处均命中。

---

### Task 5: backend-plan-agent — 拆分输出对齐粒度/具体性/评审头字段

- **Intent:** 让复杂后端任务的上游拆分产出良好粒度、具体可落地、字段对齐评审头
- **Covers spec:** §7
- **Granularity rationale:** 单文件，改 Working Rules 一条 + Output 一处，围绕"拆分质量"一个行为

**Files:**
- Modify: `plugins/backend-construct-plugin/agents/backend-plan-agent.md`

- [ ] **Step 1: 强化 Working Rules 第 5 条**

将：
```
5. 每个子任务都保持最小闭环，避免把多个不确定层混成一个大步骤。
```
替换为：
```
5. 每个子任务都保持最小闭环：一个子任务对应一个可验证行为、文件有界，避免把多个独立行为或多个不确定层混成一个大步骤（过大则拆、过小则并）。子任务必须具体可落地（点明行为/命中层/涉及文件），不得用“处理剩余逻辑”“其余按需补齐”这类空泛描述。
```

- [ ] **Step 2: Output "推荐任务拆分" 增加字段对齐说明**

将：
```
2. **推荐任务拆分**
   - 给出推荐的子任务列表。
   - 说明推荐顺序，以及为什么这样排序能降低返工或澄清不确定性。
```
替换为：
```
2. **推荐任务拆分**
   - 给出推荐的子任务列表。
   - 说明推荐顺序，以及为什么这样排序能降低返工或澄清不确定性。
   - 每个子任务标注：意图（一个可验证行为）、命中层、涉及文件——这些字段直接对应下游 plan 的评审头（Intent / Covers / Files），便于 `write-plans-with-construct` 生成可略读评审头。
```

- [ ] **Step 3: 验证**

Run: `grep -n "一个子任务对应一个可验证行为\|直接对应下游 plan 的评审头" plugins/backend-construct-plugin/agents/backend-plan-agent.md`
Expected: 两处均命中。

- [ ] **Step 4: Commit backend 改动（Task 4-5）**

```bash
git add plugins/backend-construct-plugin/skills/write-plans-with-construct/SKILL.md plugins/backend-construct-plugin/agents/backend-plan-agent.md
git commit -m "feat: align backend plan chain to task sizing, review header, and no-placeholder discipline"
```

---

### Task 6: 跨文件一致性走查

- **Intent:** 确认三套纪律在 4 文件中表述对齐、无矛盾、无遗漏
- **Covers spec:** §11
- **Granularity rationale:** 纯只读校验，独立收尾步骤

**Files:**
- 只读校验，无修改

- [ ] **Step 1: 三套标准在四文件的存在性走查**

Run:
```
cd /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market
echo "== brainstorming =="; grep -c "Task Decomposition Overview" plugins/betterpowers/skills/brainstorming/SKILL.md
echo "== writing-plans 粒度/评审头 =="; grep -c "one test-driven behavior slice\|review header" plugins/betterpowers/skills/writing-plans/SKILL.md
echo "== write-plans-with-construct 三套 =="; grep -c "任务粒度、评审头与防占位符\|评审头\|防占位符" plugins/backend-construct-plugin/skills/write-plans-with-construct/SKILL.md
echo "== backend-plan-agent =="; grep -c "一个可验证行为\|评审头" plugins/backend-construct-plugin/agents/backend-plan-agent.md
```
Expected: 各文件计数 ≥ 1（brainstorming ≥1、writing-plans ≥2、write-plans-with-construct ≥3、backend-plan-agent ≥2）。

- [ ] **Step 2: 确认无第三方依赖、改动仅限 4 个 markdown 文件**

Run: `git diff --stat HEAD~2..HEAD 2>/dev/null || git diff --stat`
Expected: 仅上述 4 个 markdown 文件（加本 plan/ spec 文档），无依赖或测试脚本。

- [ ] **Step 3: 最终报告**

汇报改了哪些文件、三套纪律落地点、跨插件漂移风险已在 spec §8 记录。

---

## Self-Review

**1. Spec coverage（逐节核对）:**
- §3.1 粒度 → Task 2（writing-plans sizing）+ Task 4（后端层内粒度）+ Task 5（agent 拆分粒度）✅
- §3.2 防占位符 → writing-plans 已有（保留，Task 2/3 不削弱）+ Task 4（后端补齐）+ Task 5（agent 具体性）✅
- §3.3 评审头 → Task 2（定义+模板规则）+ Task 3（模板落地）+ Task 4（后端层任务）+ Task 5（agent 字段对齐）✅
- §4 spec 分解概览 → Task 1 ✅
- §5 writing-plans → Task 2 + Task 3 ✅
- §6 write-plans-with-construct → Task 4 ✅
- §7 backend-plan-agent → Task 5 ✅
- §8 漂移风险 → Task 4 Step 1 新增小节首句已写明"betterpowers 为权威…需同步" ✅
- §9 YAGNI（不改 SDD/executing-plans、不动领域三级资产、不加层标签进 core、无依赖）→ 计划未触及，Task 6 Step 2 校验 ✅
- §11 验证 → 各 Task grep + Task 6 走查 ✅

**2. Placeholder scan:** 无 TBD/TODO；每步给出完整 old/new 文本与精确锚点。✅

**3. Type consistency:** 字段命名一致 —— 评审头统一用 `Intent / Covers spec / Files / Granularity rationale`（后端 `Covers` 对应命中层需求），三套纪律编号 §3.1/§3.2/§3.3 一致。✅

**4. Task granularity & headers:** 每个 task 对应一个可验证行为（按文件/纪律切分，无捆绑）；每个 task 均含评审头（Intent / Covers spec / Files / Granularity rationale）。✅
