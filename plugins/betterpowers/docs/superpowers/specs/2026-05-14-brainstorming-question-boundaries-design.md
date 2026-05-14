# Brainstorming 提问边界优化设计

日期：2026-05-14

## 背景

当前 `brainstorming` skill 默认倾向于通过逐步提问来澄清需求，这在需求模糊或存在多种可行实现时是有价值的。但在已有明确实现规范的项目中，这种默认行为会带来不必要的问题往返，拉长从需求到 spec 的路径，也容易让用户感受到“为了走流程而提问”。

本次优化目标不是削弱 brainstorming 的设计职责，而是同时收紧它的提问边界与中途确认频率：只有在真正缺少信息或存在真实决策分歧时才提问；如果仓库和当前对话已经提供了足够明确的规范，直接生成完整 spec 并交由用户审查，而不是为了流程把设计拆成多轮确认。

## 目标

1. 减少不必要提问，避免机械式澄清。
2. 保留在需求模糊或方案冲突时的提问能力。
3. 让 brainstorming 在“已有明确规范”的项目中更高效地直达 spec。
4. 保持现有整体流程主干不变：仍然先做上下文探索、仍然产出 spec、仍然要求用户审查后再进入后续计划阶段。
5. 取消不必要的“分段设计确认”，改为一次性生成完整 spec，并在落盘后输出概要供快速审阅。

## 非目标

1. 不改变 brainstorming 必须先于实现发生的硬门禁。
2. 不移除 spec 文档产出与用户审查环节。
3. 不引入新的实现技能或替代工作流。
4. 不把是否提问变成完全自由裁量，而是明确写成可执行规则。
5. 不要求在生成 spec 前把设计方案分段呈现并逐段获得确认。

## 边界定义

### 允许提问的正向边界

以下情况可以提问，并且应当只问推进决策所必需的问题：

1. 业务定义模糊，无法判断目标行为或成功标准。
2. 实现细节存在关键缺口，导致无法写出可信 spec。
3. 当前项目中存在互斥的多种实现方案，且无法仅凭现有规范判断应选哪一种。

### 不应提问的反向边界

以下情况不应提问：

1. 项目中已有明确实现规范，足以指导本次工作。
2. 仓库中已有稳定且一致的同类实现模式，可以直接沿用。
3. 用户在当前对话中已经明确指定做法、范围或约束。
4. 唯一的不确定性只是表述层面的“确认一下”，但不影响 spec 产出。

## “明确实现规范”的判定来源

brainstorming 在决定是否提问前，应主动检查以下三类来源。任一来源足够明确时，都可以构成“无需提问”的依据：

1. 项目文档与规范，如 `CLAUDE.md`、`AGENTS.md`、仓库设计文档、约定说明。
2. 代码库中已有的成熟实现模式，尤其是与当前需求同类、同层、同技术路径的实现。
3. 用户在当前对话中给出的明确指令或约束。

## 目标行为

### 新的高层流程

1. 先探索项目上下文。
2. 优先识别是否存在明确实现规范。
3. 如果存在明确规范，且没有真实冲突或关键缺口：
   - 跳过澄清提问。
   - 直接基于现有规范生成 spec。
   - 将 spec 写入既定位置。
   - 交由用户审查。
4. 如果不存在明确规范，或存在真实分歧：
   - 进入澄清问答。
   - 一次只问一个问题。
   - 问题只服务于消除关键不确定性。
5. 当信息已经足够时，不再把设计拆成多轮分段确认；直接生成完整 spec，写入既定位置，然后给用户一个简短概要用于快速审阅。

### 核心决策规则

可以将 skill 的提问逻辑明确为下面的判定：

- **默认原则：先尝试不提问。**
- 只有在以下任一条件成立时，才进入提问：
  - 业务目标不清楚。
  - 关键实现细节不足以形成 spec。
  - 当前项目存在互斥方案且没有足够依据选择。
- 如果现有规范已经足够支持设计，禁止为了“完整走流程”而提问。

## 对 `skills/brainstorming/SKILL.md` 的建议修改

## 1. Checklist 调整

将当前的：

- `Ask clarifying questions — one at a time, understand purpose/constraints/success criteria`

调整为强调条件触发的版本：

- `Ask clarifying questions only when needed — one at a time, and only to resolve true ambiguity, missing implementation detail, or mutually exclusive approaches`

这样可以保留提问步骤，但不再把“提问”写成无条件必经流程。

## 2. Process Flow 调整

在流程图中，将“Ask clarifying questions”前增加一个判断节点，例如：

- `Clear implementation norms available?`
- `Need clarification?`

建议逻辑：

- 如果已有明确规范且无真实分歧，直接跳到 `Propose 2-3 approaches` 或在实现规范唯一时直接进入完整 spec 生成路径。
- 如果存在模糊点或互斥方案，再进入 `Ask clarifying questions`。

为了和当前目标一致，建议增加一条更直接的路径：

- `Explore project context` → `Clear implementation norms available?` → `Write complete spec`

这条路径对应“直接生成 spec 供用户审查”的场景。

## 3. Understanding the idea 段落调整

在 `For appropriately-scoped projects, ask questions one at a time to refine the idea` 之前插入一段新的优先规则：

- First determine whether the repository and the current conversation already define a clear implementation path.
- If existing docs, stable code patterns, or explicit user instructions already answer the key design questions, do not ask clarifying questions.
- In that case, proceed directly to the complete design/spec using those norms, write it to the standard location, and present a short summary for review.

随后再保留现有提问规则作为条件分支，而不是默认动作。

## 4. Key Principles 调整

新增一条明确原则：

- **Norms first** — If the project already has clear implementation norms, follow them and skip unnecessary questions.

并把“提问”原则从形式约束升级为目的约束：

- **One question at a time when questions are necessary**

这样可以避免模型把 “one question at a time” 理解成“总要先问问题”。

## 5. 行为措辞调整

避免出现会强化“必须先问再设计”的表达。凡是当前暗示“先问问题再形成设计”的句子，都应调整为：

- 优先从现有规范中提取答案。
- 仅在规范不足时提问。
- 提问是例外分支，不是默认分支。

## 设计后的预期行为示例

### 场景 A：已有明确规范

用户说要优化某个 skill，仓库中已有清晰结构、命名方式、文档位置规范，且用户还明确说明边界。

预期行为：

- brainstorming 先读取仓库上下文与用户要求。
- 不再追问“你希望怎么组织 spec”“是否需要遵循现有结构”之类问题。
- 直接生成符合仓库规范的完整 spec。
- 写入 `docs/superpowers/specs/...`。
- 向用户输出一个简要概要，帮助快速判断方向是否正确，再进入文档审查。

### 场景 B：业务定义模糊

用户说“想让这个 skill 更智能一点”，但没有定义是减少提问、增强提问、还是调整触发条件。

预期行为：

- brainstorming 可以提问。
- 问题应集中在业务目标与成功标准，而不是泛泛而谈。

### 场景 C：存在互斥实现路径

用户要改一个 skill 的行为，但仓库内同时存在两种可能遵循的工作流，且文档未说明哪种优先。

预期行为：

- brainstorming 可以提问。
- 问题应聚焦于选型冲突本身，而不是扩散到无关细节。

## 风险与应对

### 风险 1：模型过度把“已有模式”当作“已有规范”

风险表现：实际上仍有关键决策缺口，但模型误以为可以直接写 spec。

应对：明确写出“只有在现有规范足以回答关键设计问题时，才能跳过提问”。如果 spec 的核心行为、边界或约束仍不清晰，则必须提问。

### 风险 2：模型跳过提问后，直接输出过度武断的 spec

风险表现：减少提问后，spec 可能更快，但更容易携带未经确认的假设。

应对：在规则中强调，只有“无真实冲突、无关键缺口”时才能跳过提问；否则必须问最少量的关键问题。

### 风险 3：流程图和正文不一致

风险表现：正文说可以不提问，图和 checklist 仍表现为必须提问，模型行为会摇摆。

应对：本次修改必须同时覆盖 checklist、流程图、正文规则和关键原则，保持一致。

## 验证思路

修改完成后，至少应人工验证以下三类提示：

1. **规范明确型**：给出已有规范和明确边界，确认 brainstorming 直接生成 spec，不再追问。
2. **需求模糊型**：只给出宽泛方向，确认 brainstorming 会提关键问题。
3. **方案冲突型**：构造两个互斥实现路径，确认 brainstorming 会围绕分歧本身提问。

重点观察的不是“是否还会提问”，而是：

- 该问时会不会问；
- 不该问时能不能直接产出 spec；
- 提问是否最小化且与决策直接相关。

## 对用户呈现方式的附加要求

当 brainstorming 生成并写入 spec 后，对用户的呈现应当改为：

1. 不逐段贴出完整设计内容并要求逐段确认。
2. 直接告知 spec 文件路径。
3. 输出一个简要概要，覆盖本次 spec 的核心目标、关键行为变化、主要边界和下一步审查动作。
4. 邀请用户基于文件和概要进行一次性审阅。

这样可以减少对话阻塞，同时保留正式的文档审查关口。

## 建议结论

本次优化应采用“保留现有框架、强化提问边界”的中等改动方案：

- 不推翻 brainstorming 的核心工作流。
- 明确把“已有规范优先”写进 skill。
- 将提问从默认动作改为条件动作。
- 当规范已足够时，直接生成 spec 并交由用户审查。

这能在不大幅扰动整体行为的前提下，精确解决当前最明显的体验问题：不必要提问。