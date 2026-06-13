# Test Curation Before Merge Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) to implement this plan with fixed-role subagents for implementation, spec review, and code quality review, or superpowers:executing-plans for inline execution. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `finishing-a-development-branch` 技能里加一个合并前的测试策展闸门，让只有高质量、有代表性的测试进入基线。

**Architecture:** 纯技能文档（markdown）改动。在 `finishing-a-development-branch/SKILL.md` 现有流程中插入一个新步骤（Step 3b），并同步该技能的 Overview/Quick Reference/Common Mistakes/Red Flags；在 `test-driven-development/SKILL.md` 加一句交叉引用；同步修正 spec §6 的落点措辞。无生产代码。

**Tech Stack:** Markdown 技能文件；验证用 `git diff` 文本走查、`grep`/`Read` 一致性检查、`bash -n` 校验抽取的 shell 片段。

**实现说明（落点精化）:** spec §6 写的是 Step 1b，但策展依赖 `TARGET_REPO`（Step 2 解析）与 `BASE_BRANCH`（Step 3 解析）。本计划把落点精化为 **Step 3b**（Step 3 之后、Step 4 之前），并在 Task 4 同步修正 spec §6，使二者一致。

---

### Task 1: 在 finishing-a-development-branch 插入 Step 3b 策展步骤

**Files:**
- Modify: `plugins/betterpowers/skills/finishing-a-development-branch/SKILL.md`（在 Step 3 与 Step 4 之间插入）

- [ ] **Step 1: 确认插入锚点存在**

Run: `grep -n "### Step 4: Present Options" plugins/betterpowers/skills/finishing-a-development-branch/SKILL.md`
Expected: 命中一行（当前约在第 84 行），作为插入位置的下界锚点。

- [ ] **Step 2: 在 `### Step 4: Present Options` 之前插入以下整段**

在 `### Step 4: Present Options` 这一行的正上方插入（保留其与上文之间的空行）：

````markdown
### Step 3b: Curate the Branch's Tests

**Runs after the base/output branches are known (Step 3) and before presenting options (Step 4). Requires tests already green from Step 1.**

The TDD cycle accumulates drop-down diagnostic tests that were useful during development but should not all persist into the baseline regression suite. Before this branch's work enters the baseline, curate the tests it added so only high-quality, representative tests remain.

**Scope:** Only tests this branch added or modified — never pre-existing baseline tests.

```bash
# BASE_BRANCH was resolved in Step 3. Diff only test files this branch touched.
git -C "$TARGET_REPO" diff --name-only "$BASE_BRANCH"...HEAD -- '*test*' '*spec*'
```

If no test files changed: report "No tests added on this branch — skipping curation." and continue to Step 4.

**Classify each added/modified test:**

| Class | Criteria | Action |
|-------|----------|--------|
| KEEP | E2E/integration proving a real workflow; the sole coverage for a behavior with no natural E2E path; a smaller test guarding a distinct failure mode or fixed-bug regression not fully covered by a retained larger test | Leave untouched |
| AUTO-DELETE | Drop-down/diagnostic test whose behavior is now fully exercised by a retained larger test; trivial happy-path unit duplicating an E2E; internal-helper test with no independent regression value | Delete without asking; list in report |
| ASK | Overlaps a larger test but may touch an edge it misses; slow test of unclear value; any case where you are not confident coverage is preserved after deletion | List with rationale + recommendation; wait for the user |

**Safety invariants (MANDATORY):**

1. **Coverage preservation:** Never auto-delete a test if doing so leaves a behavior with no retained coverage. If unsure coverage is preserved, downgrade it to ASK.
2. **Re-run after deletion:** After deletions, re-run the retained suite. It must pass. If a new failure appears (the deleted test was load-bearing — e.g. shared setup), restore that test and reclassify.
3. **Separate commit:** Commit deletions on their own with a clear message, e.g. `test: curate <feature> suite — remove N diagnostic tests subsumed by E2E`.

**Then print the curation report and continue to Step 4:**

```
Test curation (vs <base-branch>):
  Kept (N):
    - <test> — <representative reason>
  Deleted (M):
    - <test> — <why redundant>
  Need your call (K):        # only if any ASK items
    - <test> — <overlap/uncertainty> — recommend keep/delete
Retained suite: <pass/fail> (runtime <before> → <after> if available)
```

````

- [ ] **Step 3: 验证插入正确且未破坏 Step 4**

Run: `grep -n "### Step 3b: Curate the Branch's Tests\|### Step 4: Present Options" plugins/betterpowers/skills/finishing-a-development-branch/SKILL.md`
Expected: 两行依次出现，Step 3b 在 Step 4 之前。

- [ ] **Step 4: 抽取并校验内嵌 shell 片段语法**

把 Step 3b 中那段 `git diff ...` 的 bash 代码块存到临时文件后校验：
Run: `printf '%s\n' 'BASE_BRANCH=main' 'TARGET_REPO=.' 'git -C "$TARGET_REPO" diff --name-only "$BASE_BRANCH"...HEAD -- "*test*" "*spec*"' | bash -n && echo OK`
Expected: 输出 `OK`（语法合法，不实际执行）。

- [ ] **Step 5: Commit**

```bash
git add plugins/betterpowers/skills/finishing-a-development-branch/SKILL.md
git commit -m "feat: add pre-merge test curation gate to finishing-a-development-branch"
```

---

### Task 2: 同步 finishing-a-development-branch 的 Overview / Quick Reference / Common Mistakes / Red Flags

**Files:**
- Modify: `plugins/betterpowers/skills/finishing-a-development-branch/SKILL.md`

- [ ] **Step 1: 更新 Overview 的 Core principle**

把这一行：
```
**Core principle:** Verify tests → Resolve target repository → Detect environment → Present options → Execute choice → Clean up.
```
改为：
```
**Core principle:** Verify tests → Resolve target repository → Detect environment → Determine branches → Curate branch tests → Present options → Execute choice → Clean up.
```

- [ ] **Step 2: 在 Common Mistakes 末尾追加三条**

在 `## Common Mistakes` 区块（最后一条 `**Cleaning up harness-owned worktrees**` 之后）追加：

````markdown
**Deleting a test without checking coverage is preserved**
- **Problem:** Removing the only test that guards a behavior silently drops regression coverage
- **Fix:** Before auto-deleting, confirm a retained test still covers the behavior; if unsure, downgrade to ASK

**Not re-running the suite after curation deletions**
- **Problem:** A deleted test may have been load-bearing (shared setup/fixtures), leaving the suite broken
- **Fix:** Re-run the retained suite after deletions; restore and reclassify if a new failure appears

**Curating pre-existing baseline tests**
- **Problem:** The gate is meant to stop new bloat, not rewrite history; touching baseline tests is out of scope and risky
- **Fix:** Only curate tests this branch added or modified (diff vs base branch)
````

- [ ] **Step 3: 在 Red Flags 的 Never / Always 列表各加一条**

`**Never:**` 列表追加：
```
- Auto-delete a test without confirming its behavior stays covered by a retained test
- Curate (delete) pre-existing baseline tests — only this branch's added/modified tests are in scope
```
`**Always:**` 列表追加：
```
- Re-run the retained suite after curation deletions and confirm it passes
```

- [ ] **Step 4: 验证一致性**

Run: `grep -n "Curate branch tests\|Curating pre-existing\|Re-run the retained suite" plugins/betterpowers/skills/finishing-a-development-branch/SKILL.md`
Expected: Core principle 行、Common Mistakes 条目、Red Flags 条目均命中。

- [ ] **Step 5: Commit**

```bash
git add plugins/betterpowers/skills/finishing-a-development-branch/SKILL.md
git commit -m "docs: sync curation gate into overview, common mistakes, and red flags"
```

---

### Task 3: 在 TDD 技能加一句指向策展闸门的交叉引用

**Files:**
- Modify: `plugins/betterpowers/skills/test-driven-development/SKILL.md`（REFACTOR 小节）

- [ ] **Step 1: 定位 REFACTOR 小节锚点**

Run: `grep -n "Do not backfill unit tests for every new function by default" plugins/betterpowers/skills/test-driven-development/SKILL.md`
Expected: 命中一行（当前约第 164 行）。

- [ ] **Step 2: 在该句所在段落末尾追加一句交叉引用**

把这一段：
```
Do not backfill unit tests for every new function by default. Add smaller tests only where they continue to carry diagnostic or regression value.
```
改为（在末尾追加一句）：
```
Do not backfill unit tests for every new function by default. Add smaller tests only where they continue to carry diagnostic or regression value. Drop-down diagnostic tests you write here do not need to be kept long-term — the superpowers:finishing-a-development-branch curation gate decides what enters the baseline before merge, so don't keep a test out of fear of deleting it.
```

- [ ] **Step 3: 验证**

Run: `grep -n "finishing-a-development-branch curation gate" plugins/betterpowers/skills/test-driven-development/SKILL.md`
Expected: 命中一行。

- [ ] **Step 4: Commit**

```bash
git add plugins/betterpowers/skills/test-driven-development/SKILL.md
git commit -m "docs: cross-reference curation gate from TDD refactor step"
```

---

### Task 4: 同步修正 spec §6 的落点措辞（Step 1b → Step 3b）

**Files:**
- Modify: `plugins/betterpowers/docs/superpowers/specs/2026-06-13-test-curation-before-merge-design.md`（§6）

- [ ] **Step 1: 定位 §6 中的落点描述**

Run: `grep -n "Step 1b\|Step 1 Verify Tests（测试已全绿）之后\|Step 2（解析仓库" plugins/betterpowers/docs/superpowers/specs/2026-06-13-test-curation-before-merge-design.md`
Expected: 命中 §6 中关于落点的几处。

- [ ] **Step 2: 把落点从 Step 1b 改为 Step 3b**

将 §6 开头：
```
在现有流程里插入一个新步骤 **Step 1b：Curate the branch's tests**，位于：
- **Step 1 Verify Tests（测试已全绿）之后** —— 只有套件本就是绿的才谈得上策展
- **Step 2（解析仓库 / 呈现选项）之前** —— 保证 merge/PR 携带的是策展后的套件
```
改为：
```
在现有流程里插入一个新步骤 **Step 3b：Curate the branch's tests**，位于：
- **Step 1 Verify Tests（测试已全绿）之后** —— 只有套件本就是绿的才谈得上策展
- **Step 3 确定 base/output 分支之后** —— 这样 `TARGET_REPO` 与 `BASE_BRANCH` 都已解析，diff 才能算
- **Step 4（呈现 merge/PR 选项）之前** —— 保证 merge/PR 携带的是策展后的套件
```
并把同节流程框里的标题 `Step 1b: Curate the branch's tests` 改为 `Step 3b: Curate the branch's tests`，把末尾 `→ 进入 Step 2` 两处改为 `→ 进入 Step 4`。

- [ ] **Step 3: 验证 spec 内不再残留 Step 1b**

Run: `grep -n "Step 1b" plugins/betterpowers/docs/superpowers/specs/2026-06-13-test-curation-before-merge-design.md; echo "exit=$?"`
Expected: 无命中（grep 退出码非 0），即 `Step 1b` 已全部替换。

- [ ] **Step 4: Commit**

```bash
git add plugins/betterpowers/docs/superpowers/specs/2026-06-13-test-curation-before-merge-design.md
git commit -m "docs: align spec curation placement to Step 3b"
```

---

### Task 5: 全局一致性走查

**Files:**
- 只读校验，无修改

- [ ] **Step 1: 确认两个技能文件互相引用且无矛盾**

Run: `grep -rn "curation\|Curate\|finishing-a-development-branch curation" plugins/betterpowers/skills/finishing-a-development-branch/SKILL.md plugins/betterpowers/skills/test-driven-development/SKILL.md`
Expected: finishing 技能含 Step 3b + Overview/Red Flags 条目；TDD 含交叉引用句；无残留 `Step 1b`/`Step 2` 误指。

- [ ] **Step 2: 确认未引入第三方依赖、未新增调真实模型的测试**

Run: `git -C plugins/betterpowers diff --stat HEAD~4..HEAD 2>/dev/null || git diff --stat`
Expected: 改动仅限上述 markdown 文件，无新增依赖或测试脚本。

- [ ] **Step 3: 最终报告**

汇报改了哪些文件、Step 3b 的位置、以及"只防未来膨胀、不清理存量"这一已知边界仍成立。

---

## Self-Review

**Spec coverage（逐节核对）:**
- §2 决策（删除 / 合并前 / 自动删+可疑才问）→ Task 1 的分类表与安全网覆盖 ✅
- §3 范围（只 diff 本分支测试、不碰存量、已知限制）→ Task 1 scope 行 + Task 2 Common Mistakes/Red Flags "out of scope" 条 ✅
- §4 三分类标准 → Task 1 分类表逐条对应 ✅
- §5 安全不变式（覆盖保全 / 复跑 / 独立 commit / 空操作短路）→ Task 1 Safety invariants + "skipping curation" 分支 ✅
- §6 落点 → Task 1 插入 + Task 4 同步 spec（落点精化为 Step 3b，已在计划顶部说明）✅
- §7 TDD 交叉引用 → Task 3 ✅
- §8 YAGNI（不做分层/存量清理/改其它流程/第三方依赖）→ 计划未触及这些，Task 5 Step 2 校验 ✅
- §9 验证方式 → 各 Task 的 grep/bash -n + Task 5 走查 ✅

**Placeholder scan:** 无 TBD/TODO；每个编辑步骤都给出了要插入/替换的完整文本与精确锚点。✅

**Type consistency:** 全程使用一致命名 —— 步骤名 `Step 3b`、`TARGET_REPO`、`BASE_BRANCH`、分类 `KEEP/AUTO-DELETE/ASK`，与 spec 一致。✅
