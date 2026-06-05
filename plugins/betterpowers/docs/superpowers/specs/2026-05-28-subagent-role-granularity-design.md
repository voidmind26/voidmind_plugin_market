# Subagent-Driven Development 职责粒度改造设计

**日期：** 2026-05-28  
**状态：** Draft

## 问题

当前 `subagent-driven-development` skill 的核心执行模型是 **task-granularity**：controller 按任务逐个派发全新的 implementer / spec-reviewer / code-quality-reviewer subagent，然后在每个任务完成后做两阶段评审。

这种模式的优点是上下文隔离强，但在连续执行一个较长 implementation plan 时也带来明显问题：

1. **角色身份不稳定**：同一个实现职责会在不同任务间反复由新的 implementer 接手，长期约束、风格和边界意识难以积累。
2. **重复上下文灌输**：controller 需要在每个任务重新解释工作方式、历史决策和当前阶段约束，通信成本高。
3. **任务边界强于职责边界**：当前模型围绕“Task N 做完了吗”运转，而不是围绕“谁负责实现、谁负责规格符合性、谁负责质量把关”运转，不利于形成稳定分工。
4. **技能叙事与角色导向不一致**：skill 已经有 implementer、spec-reviewer、code-quality-reviewer 三类明确角色，但主文案仍强调 `fresh subagent per task`，导致“角色存在但不持久”。

用户希望将该 skill 从 **按任务粒度创建 subagent** 调整为 **按职责粒度创建固定角色 subagent**：在一个 plan 执行周期内，固定少量角色 agent 持续工作，而不是每个任务重新创建整套 agent。

## 目标

1. 将 `subagent-driven-development` 的主执行模型从“每任务创建 agent”改为“每职责创建 agent”。
2. 固定三类核心职责角色：`implementer`、`spec-reviewer`、`code-quality-reviewer`。
3. 保留当前两阶段评审顺序：**spec compliance first, code quality second**。
4. 保留 controller 主导的 plan 执行、任务推进、review loop、最终收口。
5. 减少重复提示、重复上下文搬运和每任务重新建 agent 的开销。
6. 保持 skill 的高约束风格：明确角色边界、明确何时返工、明确何时升级为人工介入。

## 非目标

1. 不引入新的长期角色，如独立 `integrator` 或 `test-auditor`。
2. 不把执行模型改成按模块/业务域 owner 划分。
3. 不取消按任务/阶段推进的计划结构；变的是 **agent 生命周期**，不是 plan 必须放弃 task 结构。
4. 不取消最终全局 review 和 `finishing-a-development-branch` 收尾流程。
5. 不要求所有 harness 都支持真正并行常驻 agent；只要求 workflow 语义升级为固定职责角色。

## 设计原则

### 1. 角色持久化，任务流动化

任务仍然是 plan 的推进单位，但 agent 不再围绕任务创建和销毁。任务在固定角色之间流动，角色对其职责持续负责。

### 2. 职责边界先于任务边界

controller 负责调度与收口，implementer 负责实现，spec-reviewer 负责需求符合性，code-quality-reviewer 负责工程质量。每个角色跨多个任务保持同一种判断视角。

### 3. 评审顺序不变

无论 agent 是否持久化，都必须坚持：

1. implementer 完成当前任务/阶段
2. spec-reviewer 先验证“做得对不对”
3. code-quality-reviewer 再验证“做得好不好”

### 4. 历史可见，但范围仍要收紧

固定角色会天然拥有更多历史上下文，但 review scope 仍必须默认收紧到 **当前 task/phase requirement + 当前 diff range**。持久化不是放宽 review scope 的理由。

### 5. controller 仍是唯一流程裁决者

固定角色不意味着自主协商式多 agent 群体。controller 仍负责：

- 决定当前推进哪个 task/phase
- 决定给各角色发送哪些上下文
- 解释 reviewer 反馈是否需要返工
- 判断何时 re-dispatch、何时升级、何时向用户求助

## 新的高层模型

### 当前模型

对每个 task：
- 新建 implementer
- 新建 spec-reviewer
- 新建 code-quality-reviewer
- 完成后这一轮 agent 生命周期结束

### 新模型

在 plan 执行开始时，一次性建立固定角色：
- `implementer`
- `spec-reviewer`
- `code-quality-reviewer`

之后每个 task/phase：
- controller 把当前任务要求、增量上下文和 diff 信息发送给对应固定角色
- 同一个角色 agent 连续处理多个 task/phase
- 直到整个 plan 执行完毕才结束该角色生命周期

## 角色定义

### 1. Controller

**职责：**
- 读取 plan 并提取任务
- 创建任务跟踪
- 记录 base/head SHA
- 决定当前 task/phase 的上下文输入
- 给 implementer / reviewers 分发当前工作
- 汇总 reviewer 反馈并驱动返工
- 在必要时升级模型、拆分任务或向用户升级
- 触发最终全局 review 与收尾 skill

**不负责：**
- 直接替 implementer 写实现代码
- 将 reviewer 的建议自动视为真理

### 2. Implementer

**职责：**
- 连续实现 plan 中的多个 task/phase
- 执行测试、验证、自审
- 根据 reviewer 反馈修改实现
- 对当前任务结果给出 DONE / DONE_WITH_CONCERNS / BLOCKED / NEEDS_CONTEXT 状态

**新的关键要求：**
- implementer 不再被视为一次性 disposable subagent，而是整个 execution run 的长期实现角色
- 需要保持对已完成任务的已知约束、命名风格、局部架构选择的连续理解
- 但仍不得擅自扩展 scope，也不得因为“我记得之前提过”就跳过当前任务需求核对

### 3. Spec Reviewer

**职责：**
- 连续审查每个 task/phase 是否严格满足 specification
- 关注是否遗漏、误解、超做
- 对 pre-existing issue 继续采用 out-of-scope 处理规则

**新的关键要求：**
- 持久化后，spec-reviewer 可以保留对 earlier accepted decisions 的记忆
- 但每次 review 仍必须以“当前 task/phase requirement + 当前 diff”为主视角
- 不得因为长期参与就把 review 扩成“重新审全分支”

### 4. Code Quality Reviewer

**职责：**
- 连续审查每个 task/phase 的代码质量、测试质量、结构质量
- 关注职责边界、文件大小、接口清晰度、可维护性

**新的关键要求：**
- 长期参与后可追踪“质量债务是否在扩大”
- 但只在当前 diff 对该问题有贡献时将其作为 in-scope issue

## 新的流程设计

### Step 1：启动固定角色

controller 在读取 plan 后，不是为 Task 1 立刻创建一次性 implementer，而是先建立固定角色编制：

1. 创建 `implementer`
2. 创建 `spec-reviewer`
3. 创建 `code-quality-reviewer`
4. 向每个角色发送初始化消息，说明：
   - 它在整个 execution run 中的长期职责
   - plan 的总体目标
   - review scope 的默认规则
   - controller 是唯一调度者

这一步的核心变化是：**先建立团队角色，再分发具体任务。**

### Step 2：按 task/phase 推进，但不重建角色

对于每个 task/phase：

1. controller 标记当前 task/phase 为 in_progress
2. 记录当前 task/phase base SHA
3. 将当前任务全文、必要上下文、历史决策增量发送给固定 `implementer`
4. implementer 实现、测试、自审、提交并汇报状态
5. controller 记录 head SHA
6. 将当前 task/phase requirements + diff range 发送给固定 `spec-reviewer`
7. 若 spec-review 不通过，controller 将问题汇总回固定 `implementer`，修复后重新走 spec-review loop
8. spec-review 通过后，将当前 diff range 发送给固定 `code-quality-reviewer`
9. 若 code-quality-review 不通过，controller 将问题汇总回固定 `implementer`，修复后重新走 quality-review loop
10. 当前 task/phase 通过后，controller 标记完成并推进下一个 task/phase

### Step 3：全局收口

当所有 task/phase 完成后：

1. controller 触发最终全局 code review
2. 通过后调用 `superpowers:finishing-a-development-branch`
3. 清理整个 execution run 的角色生命周期

## 流程图调整方向

`skills/subagent-driven-development/SKILL.md` 中现有流程图以 `Per Task` 为中心，并在图中多次出现：
- `Dispatch implementer subagent`
- `Implementer subagent fixes ...`
- `Dispatch spec reviewer ...`
- `Dispatch code quality reviewer ...`

建议改为两层模型：

### 第一层：角色初始化
- Read plan / extract tasks / create task tracking
- Start persistent implementer
- Start persistent spec-reviewer
- Start persistent code-quality-reviewer

### 第二层：每任务循环
- Send current task to implementer
- Receive implementer status
- Send current diff to spec-reviewer
- If issues: route back to implementer
- Send current diff to code-quality-reviewer
- If issues: route back to implementer
- Mark task complete

这样流程语义从：
- **dispatch fresh agent for task**
变为：
- **route current task through fixed role agents**

## 对技能文本的具体修改方向

### 1. Frontmatter description

当前 description：

> Use when you have a written implementation plan, the tasks are mostly independent, and you want to execute it task-by-task in the current session with subagents

建议改为强调固定职责角色，而不是 task-by-task：

> Use when you have a written implementation plan and want to execute it in the current session using fixed-role subagents for implementation, spec review, and code quality review

### 2. Overview

当前：

> Execute plan by dispatching fresh subagent per task, with two-stage review after each

建议改为：

> Execute a written plan using fixed-role subagents: one persistent implementer, one persistent spec reviewer, and one persistent code quality reviewer, with two-stage review after each task or phase

### 3. Core principle

当前：

> Fresh subagent per task + two-stage review (spec then quality) = high quality, fast iteration

建议改为：

> Fixed role ownership + task-scoped review loops (spec then quality) = consistent execution, clearer accountability, and high quality

### 4. When to Use

流程图中的 `Tasks mostly independent?` 需要重新表述为更贴近职责粒度的判断：

- plan 是否能按 task/phase 顺序推进
- 是否适合由固定 implementer 持续执行
- 是否适合由固定 reviewer 角色持续审查

不再把 “每任务都应 fresh” 视作前提。

### 5. Phase Acceptance Rules

保留，但需补充一句：

- Persistent reviewers may remember earlier phases, but must still review the current task or phase against its own requirements and diff by default.

### 6. Handling Implementer Status

保留现有四种状态，但将措辞从“一次性 implementer subagent”改为“persistent implementer role”。

### 7. Prompt Templates

三个 prompt 模板都要从“单次 dispatch prompt”升级为“长期职责 prompt”。

## Prompt Template 改造

### `implementer-prompt.md`

当前模板默认是一次性 task dispatch：
- `You are implementing Task N`
- `When done, report ...`

建议改造为：
- 你是本 execution run 的 persistent implementer
- 你会连续收到多个 task/phase
- 对每个 task/phase 单独汇报状态
- 保持跨任务的实现连续性，但永远以当前 task 要求为准
- reviewer 的反馈将由 controller 回流给你继续修复

需要新增的约束：
- 不因记忆上个任务而越界提前做下个任务
- 不因长期角色而私自重构全局

### `spec-reviewer-prompt.md`

建议改造为：
- 你是本 execution run 的 persistent spec compliance reviewer
- 你会连续收到多个 task/phase 的 requirements 与 diff
- 你可以记住 earlier accepted decisions，但默认 review scope 只看当前 task/phase
- 除非 earlier issue 使 current task invalid，否则不要重开 earlier accepted work

### `code-quality-reviewer-prompt.md`

建议改造为：
- 你是本 execution run 的 persistent code quality reviewer
- 你会跨多个 task/phase 连续工作
- 你可以追踪质量趋势，但只有当前 diff 对该问题有贡献时才作为 in-scope issue

## 对 `writing-plans` 的联动修改

`skills/writing-plans/SKILL.md` 当前在多个位置写死了：

- `implement this plan task-by-task`
- `fresh subagent per task`

这些需要改成与职责粒度兼容的表述，例如：

- header 中改为：
  > Use superpowers:subagent-driven-development (recommended) to implement this plan with fixed-role subagents, or superpowers:executing-plans for inline execution.

- execution handoff 中改为：
  > Subagent-Driven (recommended) - I use fixed-role subagents for implementation, spec review, and code quality review, with review loops between tasks

计划本身仍然可以保留 task-based structure；变化的是执行叙事，而不是 plan 文档必须改成 role-based 目录结构。

## 测试改造

### 1. 单元/文本测试

`tests/claude-code/test-subagent-driven-development.sh` 当前断言大量依赖旧口径：
- `load plan`
- `spec compliance before code quality`
- `self-review`
- `read plan once`
- `full task text`

这些大部分可以保留，但需要把以下断言升级：

新增应验证：
1. skill 描述中提到固定职责角色
2. workflow 描述中强调 persistent implementer/spec-reviewer/code-quality-reviewer
3. 回答中不再把 `fresh subagent per task` 当成核心原则

### 2. 集成测试

`tests/claude-code/test-subagent-driven-development-integration.sh` 需要从“有没有至少派发多个 subagent”升级为更贴近角色生命周期的验证：

1. skill 被正确调用
2. 存在固定角色初始化语义
3. 同名或同角色 agent 跨多个 task 被复用
4. review 顺序仍是 spec -> quality
5. reviewer 发现问题后会回流给 implementer
6. 计划执行完成后实现结果正确

如果 harness 很难直接验证“同一个 agent instance 被复用”，则至少要验证：
- 输出和提示中明确建立 persistent role model
- controller 在后续 task 中继续以相同 role 名称通信

### 3. 兼容性注意

如果某些 harness 不支持长期后台 agent，skill 文本需要说明：
- 优先以固定角色语义实现
- 若平台技术限制无法真正保持 agent 常驻，也应尽量保持角色名、角色上下文和职责连续性
- 不能退回“每任务重新创建一组匿名 agent”作为默认叙事

## 风险与应对

### 风险 1：持久化角色导致上下文污染累积

**表现：** implementer 或 reviewer 跨多个任务后，把 earlier assumptions 错带到 later tasks。

**应对：**
- 明确规定每轮仍以当前 task/phase requirement 为准
- reviewer 仍默认只审当前 diff
- controller 每轮传入当前任务全文与必要增量上下文，避免角色自己脑补

### 风险 2：持久化 reviewer 倾向重复追旧账

**表现：** reviewer 因为记得 earlier issues，不断把已关闭问题重新拉回当前 review。

**应对：**
- 强化 out-of-scope 规则
- 明确“记住历史 != 扩大当前 review scope”

### 风险 3：实现语义升级，但测试仍锁定旧文案

**表现：** skill 已改成职责粒度，但测试仍要求 `fresh subagent per task`。

**应对：**
- 同步更新所有技能测试、集成测试、示例和 prompt 文案
- 搜索并替换 `fresh subagent per task`、`task-by-task` 等核心短语

### 风险 4：计划文档与执行模型脱节

**表现：** `writing-plans` 继续输出旧 handoff 文案，导致上游计划仍在教 agent 按 task 创建新 subagent。

**应对：**
- 将 `writing-plans` 中所有 subagent handoff 叙事同步改为 fixed-role model

## 验证思路

至少验证以下三类场景：

### 场景 A：两任务顺序实现

- Task 1 与 Task 2 由同一个 implementer 连续执行
- 两个 reviewer 角色也连续存在
- review 顺序保持不变

### 场景 B：Task 2 暴露 Task 1 的历史问题

- reviewer 能识别该问题是否由当前 diff 引入
- 若不是当前 diff 引入，则作为 out-of-scope observation，而不是直接 fail 当前 review

### 场景 C：平台不支持真正常驻 agent

- controller 仍以固定角色语义运行
- 角色命名、职责和上下文连续性仍然可见
- 不退回旧的“匿名 per-task agent”叙事

## 涉及文件

### 主要修改
- `skills/subagent-driven-development/SKILL.md`
- `skills/subagent-driven-development/implementer-prompt.md`
- `skills/subagent-driven-development/spec-reviewer-prompt.md`
- `skills/subagent-driven-development/code-quality-reviewer-prompt.md`
- `skills/writing-plans/SKILL.md`

### 测试修改
- `tests/claude-code/test-subagent-driven-development.sh`
- `tests/claude-code/test-subagent-driven-development-integration.sh`

### 可能需要同步搜索替换的文档
- `docs/superpowers/plans/*.md` 中的 handoff 示例
- 任何包含 `fresh subagent per task`、`task-by-task`、`dispatch implementer subagent` 的说明

## 结论

本次改造的核心不是放弃 task-based plan，而是把 subagent 的生命周期从任务绑定，改为职责绑定。

最终目标模型是：

- 一个固定 `implementer`
- 一个固定 `spec-reviewer`
- 一个固定 `code-quality-reviewer`
- 一个 controller 负责整个 execution run 的调度、返工和收口

这样既保留当前 `subagent-driven-development` 的高约束 review loop，也让多任务执行过程具备更稳定的职责边界、更低的重复上下文成本，以及更清晰的执行责任归属。
