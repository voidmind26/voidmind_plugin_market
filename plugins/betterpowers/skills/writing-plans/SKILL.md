---
name: writing-plans
description: Use when you have approved requirements or a spec and need a step-by-step implementation plan before touching code
---

# Writing Plans

## Overview

Write comprehensive implementation plans assuming the engineer has zero context for our codebase and questionable taste. Document everything they need to know: which files to touch for each task, code, testing, docs they might need to check, how to test it. Give them the whole plan as bite-sized tasks. DRY. YAGNI. TDD. Frequent commits.

Assume they are a skilled developer, but know almost nothing about our toolset or problem domain. Assume they don't know good test design very well.

**Announce at start:** "I'm using the writing-plans skill to create the implementation plan."

**Context:** If working in an isolated worktree, it should have been created via the `superpowers:using-git-worktrees` skill at execution time.

**Save plans to:** `docs/superpowers/plans/YYYY-MM-DD-<feature-name>.md`
- (User preferences for plan location override this default)

## Scope Check

If the spec covers multiple independent subsystems, it should have been broken into sub-project specs during brainstorming. If it wasn't, suggest breaking this into separate plans — one per subsystem. Each plan should produce working, testable software on its own.

## File Structure

Before defining tasks, map out which files will be created or modified and what each one is responsible for. This is where decomposition decisions get locked in.

- Design units with clear boundaries and well-defined interfaces. Each file should have one clear responsibility.
- You reason best about code you can hold in context at once, and your edits are more reliable when files are focused. Prefer smaller, focused files over large ones that do too much.
- Files that change together should live together. Split by responsibility, not by technical layer.
- In existing codebases, follow established patterns. If the codebase uses large files, don't unilaterally restructure - but if a file you're modifying has grown unwieldy, including a split in the plan is reasonable.

This structure informs the task decomposition. Each task should produce self-contained changes that make sense independently.

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

## Bite-Sized Task Granularity

**Each step is one action (2-5 minutes):**
- "Write the failing test" - step
- "Run it to make sure it fails" - step
- "Implement the minimal code to make the test pass" - step
- "Run the tests and make sure they pass" - step
- "Commit" - step

## Plan Document Header

**Every plan MUST start with this header:**

```markdown
# [Feature Name] Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) to implement this plan with fixed-role subagents for implementation, spec review, and code quality review, or superpowers:executing-plans for inline execution. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** [One sentence describing what this builds]

**Architecture:** [2-3 sentences about approach]

**Tech Stack:** [Key technologies/libraries]

---
```

## Task Structure

````markdown
### Task N: [Component Name]

- **Intent:** [the one verifiable behavior this task achieves]
- **Covers spec:** [which spec section/requirement]
- **Granularity rationale:** [why this is one task, not several or half of one]

**Files:**
- Create: `exact/path/to/file.py`
- Modify: `exact/path/to/existing.py:123-145`
- Test: `tests/exact/path/to/test.py`

- [ ] **Step 1: Write the failing test**

```python
def test_specific_behavior():
    result = function(input)
    assert result == expected
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pytest tests/path/test.py::test_name -v`
Expected: FAIL with "function not defined"

- [ ] **Step 3: Write minimal implementation**

```python
def function(input):
    return expected
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pytest tests/path/test.py::test_name -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add tests/path/test.py src/path/file.py
git commit -m "feat: add specific feature"
```
````

## No Placeholders

Every step must contain the actual content an engineer needs. These are **plan failures** — never write them:
- "TBD", "TODO", "implement later", "fill in details"
- "Add appropriate error handling" / "add validation" / "handle edge cases"
- "Write tests for the above" (without actual test code)
- "Similar to Task N" (repeat the code — the engineer may be reading tasks out of order)
- Steps that describe what to do without showing how (code blocks required for code steps)
- References to types, functions, or methods not defined in any task

## Remember
- Exact file paths always
- Complete code in every step — if a step changes code, show the code
- Exact commands with expected output
- DRY, YAGNI, TDD, frequent commits

## Self-Review

After writing the complete plan, look at the spec with fresh eyes and check the plan against it. This is a checklist you run yourself — not a subagent dispatch.

**1. Spec coverage:** Skim each section/requirement in the spec. Can you point to a task that implements it? List any gaps.

**2. Placeholder scan:** Search your plan for red flags — any of the patterns from the "No Placeholders" section above. Fix them.

**3. Type consistency:** Do the types, method signatures, and property names you used in later tasks match what you defined in earlier tasks? A function called `clearLayers()` in Task 3 but `clearFullLayers()` in Task 7 is a bug.

**4. Task granularity & headers:** Does each task map to exactly one verifiable behavior (no task bundles several; no step promoted to a task)? Does every task have a complete review header (Intent / Covers spec / Files / Granularity rationale)?

If you find issues, fix them inline. No need to re-review — just fix and move on. If you find a spec requirement with no task, add the task.

## Execution Handoff

After saving the plan, you MUST ask the user to choose the execution approach. Do not auto-select — even if the plan qualifies for subagent-driven development, the user decides whether the subagent overhead is justified.

Present this exact choice, filling in the task and file counts:

**"Plan complete and saved to `docs/superpowers/plans/<filename>.md`.**

**This plan has [N] tasks spanning [M] files. Choose the execution approach:**

**1. Subagent-Driven (recommended)** — I use fixed-role subagents for implementation, spec review, and code quality review, with review loops between tasks

**2. Inline Execution** — I execute all tasks inline in this session, with checkpoints for review

**Which approach?"**

Wait for the user's explicit choice before proceeding. Do not assume or default.

**If Subagent-Driven chosen:**
- **REQUIRED SUB-SKILL:** Use superpowers:subagent-driven-development
- Fixed-role subagents + two-stage review

**If Inline Execution chosen:**
- **REQUIRED SUB-SKILL:** Use superpowers:executing-plans
- Batch execution with checkpoints for review
