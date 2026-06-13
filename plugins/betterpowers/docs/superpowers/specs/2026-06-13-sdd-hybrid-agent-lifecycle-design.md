# SDD 混合代理生命周期（Hybrid Agent Lifecycle）设计

> 日期：2026-06-13
> 范围：`plugins/betterpowers/skills/subagent-driven-development/`（SKILL.md + 三个 prompt 模板）

## 1. 问题

当前 SDD 把三个角色都定义为 **persistent**（常驻）：implementer、spec reviewer、code-quality reviewer 各一个，从 plan 开始活到结束。这对 implementer 是优点（跨任务积累代码库理解），但对 reviewer 是缺陷：

- 常驻 reviewer 会逐渐积累与 implementer 相同的盲区、被其叙述锚定、产生"之前不是认可了吗"式的合理化——**共享上下文让审查退化成"自己审自己"**，削弱了两段审查的根本价值。
- 三个角色上下文都在涨，长 plan 下都有膨胀风险。

SDD 自己已经在 `SKILL.md:54` 暗示了方向（"persistent reviewers ... must still review the current task against its own requirements and diff by default"），但没彻底——reviewer 仍是常驻的。

目标：让 **implementer 保留连续性、reviewer 获得独立性**，各取所长。

## 2. 已确认的决策

| 维度 | 决策 |
|------|------|
| implementer 生命周期 | **常驻**（SendMessage 续接同一个命名代理） |
| reviewer 生命周期 | **每任务全新派发**（fresh per task），只接收本任务需求 + diff |
| 任务内复查 | **续接同一个 reviewer**（它刚指出问题，续接省解释；防的是沾染 implementer 思路，不是忘掉自己的发现） |
| 跨任务复查 | 一律全新 reviewer |
| 检查点重置（长 plan 重启 implementer） | **写入 SKILL.md 作为可选小节，默认关**，明确触发条件与交接内容 |
| 模型 | 沿用既有决策：所有子代理 sonnet，编排者 opus（本设计不改模型）。**改 prompt 模板时必须保留各 dispatch 头已有的 `model: sonnet` 标注**，不得在重写常驻→fresh 语义时丢失 |

## 3. 混合模型

### 3.1 角色生命周期

| 角色 | 生命周期 | 创建/续接 | 跨任务记忆 |
|------|----------|-----------|------------|
| Implementer | 整个 plan 一个 | plan 开始创建一次；每任务 SendMessage 续接 | 有（刻意保留连续性） |
| Spec reviewer | 每任务一个，用完即弃 | 每任务全新派发 | 无（刻意隔离） |
| Quality reviewer | 每任务一个，用完即弃 | spec 通过后全新派发 | 无（刻意隔离） |
| Final reviewer | plan 收尾一个 | 全新派发，看全量 diff | 无 |

**唯一常驻的是 implementer；所有 reviewer 都是一次性的。**

### 3.2 上下文隔离（模式的命门）

审查独立性靠**输入隔离**保证，而非靠模型大小：

- **Implementer（常驻）**：自带累积上下文 + 本任务全文 + 必要指针。
- **Spec reviewer（fresh）**：只接收 ① 本任务需求/验收标准，② 本任务 diff（base..head SHA），③ 规范审查准则。**不接收** implementer 的推理过程或会话历史。
- **Quality reviewer（fresh）**：① 本任务 diff，②（可选）本任务需求作背景，③ 质量准则。同样不接收 implementer 思路。

## 4. 单任务执行循环

```
Task N 开始
  1. controller 标记 in_progress，记录 TASK_BASE_SHA
  2. controller ──SendMessage──▶ 常驻 implementer："Task N: <全文+上下文>"
  3. implementer 实现，报告 DONE / DONE_WITH_CONCERNS / NEEDS_CONTEXT / BLOCKED
  4. controller 记录 TASK_HEAD_SHA，取 diff
  5. controller ──全新 Agent()──▶ spec reviewer { 本任务需求, diff, 准则 }
       发现问题 → SendMessage 回常驻 implementer 修
                → 【续接同一个 spec reviewer】复查（任务内续接规则）
       通过
  6. controller ──全新 Agent()──▶ quality reviewer { diff, 准则 }
       同样的 "回 implementer 修 → 续接同一个 quality reviewer 复查" 回环
       通过
  7. 跑测试（TDD）→ 标记 Task N 完成
Task N+1 → SendMessage 同一个 implementer；reviewer 重新全新派发
```

**任务内复查续接规则（明确）：** 在单个任务的审查回环内，implementer 修完缺陷后的复查**续接刚才那个 reviewer**（同一会话续接，不重新派发）。一旦进入下一个任务，reviewer 必须全新派发。

## 5. 跨任务正确性的兜底

fresh reviewer 只看局部 diff，看不到跨任务回归（如 Task 3 被 Task 7 改坏）。由三道网兜住，**全局正确性不依赖 reviewer 的长期记忆**：

1. **生产侧连续性**：常驻 implementer 把约定/决策一路带下去，源头减少漂移。
2. **机械回归网**：每任务后跑 TDD 套件（配合"合并前测试策展"基线），跨任务破坏被测试直接抓出。
3. **收尾整体审查**：plan 结束时 fresh + 全量 diff 的 final reviewer，补局部 reviewer 的全局盲区（SDD 现有 `Use a full-implementation review at the end` 已含此步，保留）。

## 6. 可选：implementer 检查点重置（默认关）

针对超长 plan，常驻 implementer 仍会膨胀。作为 SKILL.md 一个**默认关闭**的可选小节：

- **触发条件**：上下文明显退化（重复犯错、丢失早前决策、反复 NEEDS_CONTEXT），或超过约定的 K 个任务。
- **动作**：用一份**精简交接**起一个新 implementer 接班——交接内容 = 当前已完成任务摘要 + 关键决策/约定 + 相关文件清单 + 当前任务上下文。
- **默认不触发**：仅在 controller 判断必要时手动启用，避免过度设计与交接出错面。

## 7. 技能文件改动清单

### `SKILL.md`
- **Overview（line 8）**：把 "one persistent implementer, one persistent spec reviewer, and one persistent code quality reviewer" 改为反映混合生命周期（implementer 常驻；reviewer 每任务 fresh）。
- **Phase Acceptance Rules（line 54）**：把 "Persistent reviewers may remember earlier phases..." 升级为正式规则：reviewer 每任务全新派发、零跨任务记忆；任务内复查续接同一个。
- **流程图（The Process）**：把 "Start persistent spec reviewer / code quality reviewer"（plan 开始一次性创建）改为"每任务派发 fresh reviewer"。
- **Example Workflow**：更新示例，体现 implementer 用 SendMessage 续接、reviewer 每任务全新、任务内复查续接。
- **Advantages / Cost**：调整措辞——独立性来自 reviewer 隔离；Cost 里"persistent roles accumulate context"限定为仅 implementer，并指向可选检查点重置。
- **新增小节**：implementer 检查点重置（默认关，见 §6）。
- **Red Flags**：新增 —— "Never: 让 reviewer 跨任务携带上下文 / 用陈旧 reviewer 复用到新任务"；"Always: reviewer 每任务全新派发、只喂本任务需求+diff"。

> **所有模板改动的硬约束：** dispatch 头当前为 `Task tool (general-purpose, model: sonnet):`，重写生命周期语义时**原样保留 `model: sonnet`**，三个模板一个都不能丢。

### `implementer-prompt.md`
- 强化"persistent + SendMessage 续接"语义（基本保持现状，明确它是唯一常驻角色）。
- 保留 `model: sonnet`。

### `spec-reviewer-prompt.md`
- 把 "persistent spec compliance reviewer for this execution run" 改为 "fresh spec compliance reviewer for **this single task/phase**"。
- 去掉"逐个接收任务"的常驻措辞；保留已有的"只看本任务 diff"约束（与混合模式一致，无需推翻）。
- 加一句任务内复查续接说明（复查时续接本会话）。

### `code-quality-reviewer-prompt.md`
- 同 spec-reviewer：常驻语义改为每任务 fresh + 任务内复查续接。

## 8. 不做的事（YAGNI）

- 不引入 team / TeamCreate / 代理间对等通信（SDD 是轮辐式，前一轮讨论已结论：subagent 足够）。
- 不并行 implementer（SDD 明确禁止）。
- 不默认开启检查点重置。
- 不改模型分工、不改其它技能。
- 不引入第三方依赖。

## 9. 验证方式

遵循 fork "零 token 成本本地测试"约束，不新增调真实模型的行为测试。验证手段：
- 文本一致性走查：Overview / Phase Acceptance Rules / 流程图 / Example / Red Flags / 三个 prompt 模板对"implementer 常驻、reviewer 每任务 fresh、任务内复查续接"的表述彼此一致、无残留"persistent reviewer"矛盾措辞。
- `bash -n` 校验改动中涉及的内嵌 shell 片段（如有）。
- **`model: sonnet` 保留校验**：改动后三个 prompt 模板的 dispatch 头仍各含 `model: sonnet`（grep 计数 = 3）。
- 人工走查一个 2-3 任务示例，确认生命周期与续接规则落地正确。
