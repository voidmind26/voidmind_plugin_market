# betterpowers 工作流梳理与评估报告

> 生成日期：2026-06-13
> 范围：`plugins/betterpowers/`（superpowers 优化 fork，v5.1.5）

---

## 一、betterpowers 整体工作流梳理

betterpowers 是 superpowers 的优化 fork，本质是**一套用"技能自动触发"驱动的软件开发方法论**。它不是工具集合，而是一条强制流程。

### 启动机制（让一切自动发生）

- `hooks/hooks.json` 在 `SessionStart`（startup / clear / compact）时运行 `session-start`，把 `using-superpowers` 的全文注入到上下文里（即会话开头的那段 `<EXTREMELY_IMPORTANT>`）。
- 这段 bootstrap 是整套系统的"引信"——它强制 agent 在动手前先检查是否有技能适用。没有它，技能就只是躺在磁盘上的死文件。

### 主干流程（一条完整开发链）

```
需求 / 想法
  │
  ▼
brainstorming ──────── 把模糊想法逼成明确 spec
  │                    HARD-GATE：没批准设计不许写任何代码
  │  (产出 design，存 docs/superpowers/specs/)
  ▼
writing-plans ──────── 把 spec 拆成"零上下文 junior 也能执行"的任务清单
  │                    强调 TDD / YAGNI / DRY
  │  (产出 plan，存 docs/superpowers/plans/YYYY-MM-DD-xxx.md)
  ▼
执行（二选一）
  ├─ subagent-driven-development（有子代理时，推荐）
  │     固定三角色：implementer / spec-reviewer / code-quality-reviewer
  │     每个任务：实现 → 规范审查 → 质量审查 → 下一个，连续不停问
  └─ executing-plans（无子代理环境的降级方案）
  │
  ▼
requesting-code-review / receiving-code-review ── 派审查子代理 + 理性接收反馈
  │
  ▼
verification-before-completion ── 铁律：没有新鲜验证证据不许说"完成"
  │
  ▼
finishing-a-development-branch ── merge / PR / 保留 / 丢弃
```

### 横切技能（流程中随时被触发）

- **test-driven-development**：写实现前先写测试，默认从 E2E / 集成测试起步（这是 fork 改过的：默认"先证明完整路径"，再按需降到小测试）。
- **systematic-debugging**：遇到 bug 的铁律——"没找到根因不许改"。
- **using-git-worktrees**：隔离工作区。
- **dispatching-parallel-agents**：多个无关 bug / 子系统并行派发。
- **writing-skills**：元技能，用 TDD 思路写 / 改技能本身。

### 配套

- `tests/`：只保留**零 token 成本**的本地测试（brainstorm-server 的 ws 协议测试、codex 同步脚本测试、shell 语法检查）。CLAUDE.md 明确禁止加默认调真实模型的测试。
- `scripts/`：`bump-version.sh`、`sync-to-codex-plugin.sh`（同步到 codex 版本）。
- `.in_use/`、`assets/`、`docs/`：历史计划 / 规范存档。

---

## 二、是否有必要全部重构成中文？

**结论：没必要，而且不建议做。**

核心要理解一件事：**技能正文和 `description` 是给模型读的，不是给你读的；而对话输出语言由全局 CLAUDE.md 的"所有交互使用中文"控制，跟技能本身用什么语言无关。** 也就是说——技能保持英文，Claude 照样全程用中文交流。翻译技能并不能让交互"更中文"。

具体理由：

| 维度 | 翻译成中文的实际收益 / 代价 |
|------|------------------------------|
| 交互语言 | **零收益**。输出已经是中文，由全局偏好控制。 |
| 模型理解 | 零收益。Claude 对英文技能的理解不弱于中文。 |
| 行为调优 | **高风险**。这些是经过 eval 反复调过的"行为塑造"文本——`Red Flags` 表、`HARD-GATE`、`Iron Law`、"human partner" 措辞、合理化清单——CLAUDE.md 明确警告这些不能随意改写 / 翻译，极易丢掉触发力度。 |
| 触发匹配 | 轻微负面。`description` 字段参与技能发现匹配，与英文工具生态（Task、git、TDD 等术语）同语种匹配更稳。 |
| 维护成本 | **最高代价**。一旦翻译就永久 fork 偏离上游，`sync-to-codex-plugin.sh` 和未来 `git` 同步全部失效，上游更新再也合不进来。 |

**真正值得本地化的只有"用户能看到的那几句"**（比如 `Announce: "I'm using the X skill..."`），但这部分 Claude 本来就会按中文偏好自然渲染，不需要改文件。

> 如果执意要做，唯一安全的范围是：只把面向阅读的提示句中文化，**绝不动** Red Flags / Iron Law / HARD-GATE / 合理化清单这类调优正文。但投入产出比很低，不推荐。

---

## 三、把 SDD 改成"全部依靠 sonnet 模型实现"的好办法

### 现状

SDD 的 `SKILL.md` 里有一节 **Model Selection**（约 122–135 行），现在的策略是**让编排者按任务复杂度挑模型**——机械任务用便宜模型、集成任务用标准模型、架构 / 审查用最强模型。它**没有硬编码任何模型**，只是给编排 agent 的指导原则。模型最终通过 Task / Agent 工具的 `model` 参数在派发子代理时决定。

所以"全部用 sonnet" = 把这套"按需挑模型"的逻辑改成"固定 sonnet"。有三种做法，从轻到重：

### 方案 A：直接改 Model Selection 段（最快，改动最小）

把那节指导文字改成"三个角色一律用 sonnet"。

- ✅ 优点：一处改动，立即生效。
- ⚠️ 缺点：它只是"指导"，编排者仍是 prompt 驱动，偶尔可能不严格遵守；属于改动调优正文（这节相对偏配置，风险比 Red Flags 低）。

### 方案 B：定义三个固定 model 的 agent（最稳，声明式，推荐）

在 `plugins/betterpowers/agents/` 下新建 `sdd-implementer`、`sdd-spec-reviewer`、`sdd-code-quality-reviewer` 三个 agent 定义文件，frontmatter 里写死 `model: sonnet`，再把 SDD 的派发说明改成"用这三个 agentType"。

- ✅ 优点：模型在 agent 定义层钉死，不依赖编排者临场判断；可维护、可复用；改动集中在新增文件 + 派发指令，不碰调优正文。
- ⚠️ 缺点：要新建 3 个文件并改 SDD 的 prompt 模板引用。

### 方案 C：整会话锁 sonnet（最粗暴）

直接 `/model sonnet` 或在 settings 里设默认模型。

- ⚠️ 注意：这会把**编排者本身**也降成 sonnet，brainstorming / writing-plans 这种需要强判断的环节质量会下降。多数人想要的是"编排 + 设计用强模型、SDD 三个干活的子代理用 sonnet"，那这个方案就不合适。

### 推荐

**方案 B**。它把"sonnet"钉在 agent 定义里而不是靠 prompt 自觉，最可靠也最好维护，并且不破坏 SDD 的调优正文。如果只是想快速试一下，先用方案 A 验证效果，满意后再升级到 B。

### 待确认问题

"SDD 全部依靠 sonnet"指的是：

- **选项 1**：只把 implementer / spec-reviewer / code-quality-reviewer 这三个干活的子代理换成 sonnet（编排者和 brainstorming / writing-plans 保持当前强模型）→ 用**方案 B**。
- **选项 2**：连编排在内整条链都跑 sonnet → 用**方案 C**。

---

## 附：关键文件位置

| 内容 | 路径 |
|------|------|
| 启动引信 hook | `hooks/hooks.json` + `hooks/session-start` |
| 主干技能 | `skills/<skill-name>/SKILL.md` |
| SDD 模型选择 | `skills/subagent-driven-development/SKILL.md`（约 122–135 行） |
| SDD 三角色 prompt 模板 | `skills/subagent-driven-development/*-prompt.md` |
| 贡献 / 修改约束 | `plugins/betterpowers/CLAUDE.md` |
| 本地测试 | `tests/`（零 token 成本） |
