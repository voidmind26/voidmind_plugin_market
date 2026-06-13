# SDD Hybrid Agent Lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) to implement this plan with fixed-role subagents for implementation, spec review, and code quality review, or superpowers:executing-plans for inline execution. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 SDD 三角色的"全持久"改成混合生命周期——implementer 常驻、spec/quality reviewer 每任务全新派发、任务内复查续接同一个 reviewer。

**Architecture:** 纯技能文档（markdown）改动，集中在 `subagent-driven-development/` 一个技能的 4 个文件。无生产代码。改 reviewer prompt 模板的生命周期语义时必须保留 `model: sonnet` 标注。

**Tech Stack:** Markdown 技能文件；验证用 `grep` 文本一致性 + `model: sonnet` 计数校验。

**关键约束:** 三个 prompt 模板 dispatch 头当前为 `Task tool (general-purpose, model: sonnet):`，所有改动**原样保留 `model: sonnet`**，执行末尾用 grep 计数=3 兜底。

---

### Task 1: SKILL.md — Overview 与 Phase Acceptance Rules

**Files:**
- Modify: `plugins/betterpowers/skills/subagent-driven-development/SKILL.md`（line 8、line 54）

- [ ] **Step 1: 改写 Overview 行（line 8）**

将：
```
Execute a written plan using fixed-role subagents: one persistent implementer, one persistent spec reviewer, and one persistent code quality reviewer, with two-stage review after each task or phase.
```
替换为：
```
Execute a written plan using fixed-role subagents with a hybrid lifecycle: one persistent implementer (continued across tasks via SendMessage to preserve codebase continuity), plus a fresh spec reviewer and a fresh code quality reviewer dispatched per task (no cross-task memory, so each review stays independent), with two-stage review after each task or phase.
```

- [ ] **Step 2: 升级 Phase Acceptance Rules（line 54）**

将：
```
- Persistent reviewers may remember earlier phases, but must still review the current task or phase against its own requirements and diff by default
```
替换为：
```
- Dispatch a fresh spec reviewer and a fresh code quality reviewer for each task or phase, carrying no memory of earlier tasks — independence is the whole point of review. Within a single task's review loop, a re-check after the implementer fixes issues continues that same reviewer (it already raised the issue); each new task always gets a freshly dispatched reviewer.
```

- [ ] **Step 3: 验证**

Run: `grep -n "hybrid lifecycle\|Dispatch a fresh spec reviewer and a fresh code quality reviewer for each task" plugins/betterpowers/skills/subagent-driven-development/SKILL.md`
Expected: 两行均命中（Overview + Phase Acceptance Rules）。

---

### Task 2: SKILL.md — 流程图改成"只前置启动 implementer，reviewer 循环内每任务派发"

**Files:**
- Modify: `plugins/betterpowers/skills/subagent-driven-development/SKILL.md`（The Process 流程图，line 72-119）

- [ ] **Step 1: 删除两个 reviewer 的前置启动节点声明（line 73-74）**

将：
```
    "Start persistent implementer (./implementer-prompt.md)" [shape=box];
    "Start persistent spec reviewer (./spec-reviewer-prompt.md)" [shape=box];
    "Start persistent code quality reviewer (./code-quality-reviewer-prompt.md)" [shape=box];
    "More tasks remain?" [shape=diamond];
```
替换为：
```
    "Start persistent implementer (./implementer-prompt.md)" [shape=box];
    "More tasks remain?" [shape=diamond];
```

- [ ] **Step 2: 重写前置启动连线（line 94-97）**

将：
```
    "Read plan, extract all tasks with full text, note context, create TodoWrite" -> "Start persistent implementer (./implementer-prompt.md)";
    "Start persistent implementer (./implementer-prompt.md)" -> "Start persistent spec reviewer (./spec-reviewer-prompt.md)";
    "Start persistent spec reviewer (./spec-reviewer-prompt.md)" -> "Start persistent code quality reviewer (./code-quality-reviewer-prompt.md)";
    "Start persistent code quality reviewer (./code-quality-reviewer-prompt.md)" -> "More tasks remain?";
```
替换为：
```
    "Read plan, extract all tasks with full text, note context, create TodoWrite" -> "Start persistent implementer (./implementer-prompt.md)";
    "Start persistent implementer (./implementer-prompt.md)" -> "More tasks remain?";
```

- [ ] **Step 3: 重命名循环内 spec reviewer 派发节点（replace_all）**

把所有出现的：
```
Send current task/phase requirements + diff to spec reviewer
```
整串替换为（replace_all = true）：
```
Dispatch this task's spec reviewer (requirements + diff; fresh per task)
```

- [ ] **Step 4: 重命名循环内 quality reviewer 派发节点（replace_all）**

把所有出现的：
```
Send current diff to code quality reviewer
```
整串替换为（replace_all = true）：
```
Dispatch this task's code quality reviewer (diff; fresh per task)
```

- [ ] **Step 5: 在流程图代码块结束 `}` 之后补一句续接说明**

将（流程图闭合 + 其后第一段，line 119 的 `}` 与后续 `### vs.` 之间——实际锚点用 `}` 后紧跟的内容）：
```
}
```
（这是流程图 dot 代码块的闭合括号那一行）后紧接插入一段普通正文：
```

> **Reviewer lifecycle note:** Only the implementer is started up front. Each task dispatches its own fresh spec and code-quality reviewers. In the diagram, the loop back to a reviewer node after a fix means *continue that task's same reviewer* on the updated diff (task-internal re-check), not a brand-new dispatch — a fresh dispatch happens only when a new task begins.
```

（注意：仅在流程图那个 dot 代码块的闭合 ``` 之后插入，不要误改其它 `}`。）

- [ ] **Step 6: 验证**

Run: `grep -n "Dispatch this task's spec reviewer\|Dispatch this task's code quality reviewer\|Reviewer lifecycle note\|Start persistent spec reviewer" plugins/betterpowers/skills/subagent-driven-development/SKILL.md`
Expected: 前三者命中；`Start persistent spec reviewer` **不再命中**（已从流程图移除）。

---

### Task 3: SKILL.md — Example Workflow 更新为混合模式

**Files:**
- Modify: `plugins/betterpowers/skills/subagent-driven-development/SKILL.md`（Example Workflow 代码块）

- [ ] **Step 1: 替换启动与 Task 1/Task 2 示例**

将：
```
[Read plan file once: docs/superpowers/plans/feature-plan.md]
[Extract all tasks with full text and context]
[Create TodoWrite with all tasks]
[Start persistent implementer]
[Start persistent spec-reviewer]
[Start persistent code-quality-reviewer]

Task 1:
[Send Task 1 text + context to implementer]
[Implementer reports DONE]
[Send Task 1 requirements + diff to spec-reviewer]
[Spec-reviewer approves]
[Send Task 1 diff to code-quality-reviewer]
[Code-quality-reviewer approves]
[Mark Task 1 complete]

Task 2:
[Send Task 2 text + context to same implementer]
[Implementer reports DONE_WITH_CONCERNS]
[Send Task 2 requirements + diff to same spec-reviewer]
[Spec-reviewer finds a missing requirement]
[Route issue back to same implementer]
[Implementer fixes and reports back]
[Re-run spec review]
[Run code quality review]
[Mark Task 2 complete]
```
替换为：
```
[Read plan file once: docs/superpowers/plans/feature-plan.md]
[Extract all tasks with full text and context]
[Create TodoWrite with all tasks]
[Start persistent implementer]              # only the implementer is started up front

Task 1:
[SendMessage Task 1 text + context to implementer]
[Implementer reports DONE]
[Dispatch FRESH spec-reviewer with Task 1 requirements + diff]
[Spec-reviewer approves]
[Dispatch FRESH code-quality-reviewer with Task 1 diff]
[Code-quality-reviewer approves]
[Mark Task 1 complete — both reviewers discarded]

Task 2:
[SendMessage Task 2 text + context to SAME implementer]
[Implementer reports DONE_WITH_CONCERNS]
[Dispatch FRESH spec-reviewer with Task 2 requirements + diff]   # new agent, no memory of Task 1
[Spec-reviewer finds a missing requirement]
[Route issue back to same implementer via SendMessage]
[Implementer fixes and reports back]
[CONTINUE the same Task 2 spec-reviewer to re-check the fix]     # task-internal continue
[Spec-reviewer approves]
[Dispatch FRESH code-quality-reviewer with Task 2 diff]
[Code-quality-reviewer approves]
[Mark Task 2 complete]
```

- [ ] **Step 2: 验证**

Run: `grep -n "only the implementer is started up front\|Dispatch FRESH spec-reviewer\|CONTINUE the same Task 2 spec-reviewer" plugins/betterpowers/skills/subagent-driven-development/SKILL.md`
Expected: 三行均命中。

---

### Task 4: SKILL.md — Advantages/Cost、新增检查点重置小节、Red Flags、提交 SKILL.md

**Files:**
- Modify: `plugins/betterpowers/skills/subagent-driven-development/SKILL.md`（Efficiency gains / Cost / 新增小节 / Red Flags）

- [ ] **Step 1: 调整 Efficiency gains 末行**

将：
```
- Controller invests upfront in role setup, then routes task-scoped context through the same roles
```
替换为：
```
- Controller invests upfront in implementer setup, then routes task-scoped context through the persistent implementer while giving each task an independently dispatched reviewer
```

- [ ] **Step 2: Quality gates 增补一条独立性优势**

将：
```
- Code quality ensures implementation is well-built
```
替换为：
```
- Code quality ensures implementation is well-built
- Per-task fresh reviewers carry no accumulated bias from earlier tasks, so each review is genuinely independent
```

- [ ] **Step 3: 修正 Cost 第一条**

将：
```
- Persistent roles accumulate more local context and need tighter scope discipline
```
替换为：
```
- The persistent implementer accumulates local context and needs scope discipline (and, on very long plans, an optional checkpoint reset — see "Implementer Checkpoint Reset"); fresh per-task reviewers avoid this accumulation entirely
```

- [ ] **Step 4: 在 "## Handling Implementer Status" 小节之前插入新小节**

在 `## Handling Implementer Status` 这一行正上方插入：
```
## Implementer Checkpoint Reset (Optional — default off)

The implementer is the only role that persists across tasks, so on a very long plan its context can degrade. This reset is **off by default**; apply it only when the controller judges it necessary.

**Trigger when:** the implementer repeatedly errs, loses track of earlier decisions, or churns on NEEDS_CONTEXT — or after a pre-agreed number of tasks (K) on an unusually long plan.

**Action:** start a new implementer seeded with a compact handoff instead of the full accumulated context. The handoff contains:
- a short summary of completed tasks and their outcomes
- key decisions and conventions established so far
- the list of files touched and their responsibilities
- the current task's full text and local context

The fresh implementer takes over from the next task. Reviewers are unaffected — they are already fresh per task.

```

- [ ] **Step 5: Red Flags — Never 列表增补两条**

将：
```
- Move to next task while either review has open issues
```
替换为：
```
- Move to next task while either review has open issues
- Carry reviewer context across tasks — each task gets a freshly dispatched reviewer with no memory of earlier tasks
- Reuse a prior task's reviewer for a new task
```

- [ ] **Step 6: Red Flags — "If reviewer finds issues" 澄清续接**

将：
```
**If reviewer finds issues:**
- The same implementer role fixes them
- Reviewer reviews again
- Repeat until approved
- Don't skip the re-review
```
替换为：
```
**If reviewer finds issues:**
- The same implementer role fixes them
- The same task's reviewer (continued within this task's loop) reviews again — do not spin up a new reviewer mid-task
- Repeat until approved
- Don't skip the re-review
```

- [ ] **Step 7: Red Flags — "If the implementer role fails repeatedly" 指向重置**

将：
```
**If the implementer role fails repeatedly:**
- Upgrade or replace the role explicitly
```
替换为：
```
**If the implementer role fails repeatedly:**
- Upgrade or replace the role explicitly (or apply a checkpoint reset — see "Implementer Checkpoint Reset")
```

- [ ] **Step 8: 验证 SKILL.md 全部命中且无残留矛盾**

Run: `grep -n "Implementer Checkpoint Reset\|Per-task fresh reviewers carry no accumulated bias\|Carry reviewer context across tasks\|continued within this task's loop" plugins/betterpowers/skills/subagent-driven-development/SKILL.md`
Expected: 四处均命中。

- [ ] **Step 9: Commit SKILL.md**

```bash
git add plugins/betterpowers/skills/subagent-driven-development/SKILL.md
git commit -m "feat: hybrid agent lifecycle in SDD (persistent implementer, fresh per-task reviewers)"
```

---

### Task 5: implementer-prompt.md — 标注唯一常驻角色（保留 model: sonnet）

**Files:**
- Modify: `plugins/betterpowers/skills/subagent-driven-development/implementer-prompt.md`（line 3、line 9）

- [ ] **Step 1: 改 line 3 说明**

将：
```
Use this template when starting the persistent implementer role for an execution run.
```
替换为：
```
Use this template when starting the persistent implementer role for an execution run. The implementer is the ONLY persistent role; spec and code-quality reviewers are dispatched fresh per task.
```

- [ ] **Step 2: 改 line 9 角色自述**

将：
```
    You are the persistent implementer role for this execution run.
```
替换为：
```
    You are the persistent implementer role for this execution run — the only role that persists across tasks. The controller continues this same session for each new task.
```

- [ ] **Step 3: 验证（含 model: sonnet 仍在）**

Run: `grep -n "the only role that persists across tasks\|model: sonnet" plugins/betterpowers/skills/subagent-driven-development/implementer-prompt.md`
Expected: 两者均命中（dispatch 头的 `model: sonnet` 未被破坏）。

- [ ] **Step 4: Commit**

```bash
git add plugins/betterpowers/skills/subagent-driven-development/implementer-prompt.md
git commit -m "docs: mark implementer as the sole persistent SDD role"
```

---

### Task 6: spec-reviewer-prompt.md — 常驻语义改 fresh + 任务内续接（保留 model: sonnet）

**Files:**
- Modify: `plugins/betterpowers/skills/subagent-driven-development/spec-reviewer-prompt.md`（line 3、line 11、line 13、line 45）

- [ ] **Step 1: 改 line 3 说明**

将：
```
Use this template when starting the persistent spec compliance reviewer role for an execution run.
```
替换为：
```
Use this template to dispatch a fresh spec compliance reviewer for a single task or phase. Dispatch a new one per task — never carry it across tasks.
```

- [ ] **Step 2: 改 line 11 角色自述**

将：
```
    You are the persistent spec compliance reviewer for this execution run.
```
替换为：
```
    You are a fresh spec compliance reviewer for this single task or phase. You have no memory of earlier tasks — review only what you are given below.
```

- [ ] **Step 3: 改 line 13 接收措辞**

将：
```
    You will receive one task or phase at a time. For each review, judge the current task or phase against its own requirements and diff by default.
```
替换为：
```
    Judge this task or phase against its own requirements and diff. If the controller later sends back a fix for an issue you raised in THIS task, re-check only that fix against the updated diff and re-report.
```

- [ ] **Step 4: 改 line 45 记忆措辞**

将：
```
    You may remember earlier accepted phases, but that memory does not widen the current review scope by default.
```
替换为：
```
    You have no memory of earlier phases; review strictly within the scope given here.
```

- [ ] **Step 5: 验证（含 model: sonnet 仍在）**

Run: `grep -n "fresh spec compliance reviewer for this single task\|no memory of earlier phases; review strictly\|model: sonnet" plugins/betterpowers/skills/subagent-driven-development/spec-reviewer-prompt.md`
Expected: 三者均命中。

- [ ] **Step 6: Commit**

```bash
git add plugins/betterpowers/skills/subagent-driven-development/spec-reviewer-prompt.md
git commit -m "docs: spec reviewer is fresh per task with task-internal re-check"
```

---

### Task 7: code-quality-reviewer-prompt.md — 常驻语义改 fresh + 任务内续接（保留 model: sonnet）

**Files:**
- Modify: `plugins/betterpowers/skills/subagent-driven-development/code-quality-reviewer-prompt.md`（line 3、line 11）

- [ ] **Step 1: 改 line 3 说明**

将：
```
Use this template when starting the persistent code quality reviewer role for an execution run.
```
替换为：
```
Use this template to dispatch a fresh code quality reviewer for a single task or phase. Dispatch a new one per task — never carry it across tasks.
```

- [ ] **Step 2: 改 line 11 角色自述**

将：
```
You are the persistent code quality reviewer for this execution run. You review one task or phase at a time, keep continuity across the run, and still judge current quality issues primarily from the current diff.
```
替换为：
```
You are a fresh code quality reviewer for this single task or phase. You have no memory of earlier tasks; judge quality from the current diff only. If the controller sends back a fix for an issue you raised in THIS task, re-check only that fix against the updated diff and re-report.
```

- [ ] **Step 3: 验证（含 model: sonnet 仍在）**

Run: `grep -n "fresh code quality reviewer for this single task\|no memory of earlier tasks; judge quality\|model: sonnet" plugins/betterpowers/skills/subagent-driven-development/code-quality-reviewer-prompt.md`
Expected: 三者均命中。

- [ ] **Step 4: Commit**

```bash
git add plugins/betterpowers/skills/subagent-driven-development/code-quality-reviewer-prompt.md
git commit -m "docs: code quality reviewer is fresh per task with task-internal re-check"
```

---

### Task 8: 全局一致性走查 + model: sonnet 计数校验

**Files:**
- 只读校验，无修改

- [ ] **Step 1: 确认无残留"persistent reviewer"矛盾措辞**

Run: `grep -rn "persistent spec\|persistent code quality\|persistent spec-reviewer\|Start persistent spec\|Start persistent code quality\|same spec-reviewer\b" plugins/betterpowers/skills/subagent-driven-development/; echo "exit=$?"`
Expected: 无命中（exit 非 0）。即所有 reviewer 的"常驻"措辞已清除。

- [ ] **Step 2: model: sonnet 计数 = 3**

Run: `grep -rc "model: sonnet" plugins/betterpowers/skills/subagent-driven-development/implementer-prompt.md plugins/betterpowers/skills/subagent-driven-development/spec-reviewer-prompt.md plugins/betterpowers/skills/subagent-driven-development/code-quality-reviewer-prompt.md | grep -c ":1"`
Expected: `3`（三个模板各保留一处 `model: sonnet`）。

- [ ] **Step 3: 确认 implementer 仍为常驻、reviewer 为 fresh 的表述一致**

Run: `grep -rn "persistent implementer\|fresh spec compliance reviewer\|fresh code quality reviewer\|hybrid lifecycle" plugins/betterpowers/skills/subagent-driven-development/`
Expected: implementer 常驻、两个 reviewer fresh、Overview hybrid lifecycle 均一致出现。

- [ ] **Step 4: 最终报告**

汇报改了哪些文件、混合生命周期落地点、`model: sonnet` 计数=3、以及无残留矛盾措辞。

---

## Self-Review

**Spec coverage（逐节核对）:**
- §2 决策（implementer 常驻 / reviewer fresh / 任务内续接 / 跨任务全新 / 检查点重置默认关 / 保留 sonnet）→ Task 1,3,4,5,6,7 + Task 8 校验 ✅
- §3.1 角色生命周期表 → Task 1 Overview + Task 2 流程图 + Task 5/6/7 模板 ✅
- §3.2 上下文隔离 → Task 6/7 "no memory" 措辞 + Task 1 Overview ✅
- §4 单任务循环 + 任务内续接规则 → Task 2 流程图 note + Task 3 Example + Task 4 Red Flags 续接澄清 ✅
- §5 跨任务兜底（三道网）→ 流程图保留 final reviewer（未删）+ Phase Acceptance Rules 收尾整体审查（line 63 原文保留）+ TDD 每任务测试（implementer-prompt 原有 TDD 步骤保留）✅
- §6 检查点重置（默认关）→ Task 4 Step 4 新增小节 ✅
- §7 文件改动清单 → Task 1-7 一一对应；硬约束（保留 model: sonnet）→ Task 5/6/7 验证步 + Task 8 计数 ✅
- §8 YAGNI（不引入 team/并行/默认重置/改模型/第三方依赖）→ 计划未触及 ✅
- §9 验证方式 → 各 Task grep + Task 8 一致性/计数走查 ✅

**Placeholder scan:** 无 TBD/TODO；每个编辑步骤给出完整 old/new 文本与精确锚点；K 为可选重置的可调阈值（非占位）。✅

**Type consistency:** 命名一致 —— "persistent implementer"、"fresh spec/code quality reviewer"、"task-internal continue/re-check"、"Implementer Checkpoint Reset"、`model: sonnet` 全程统一。✅
