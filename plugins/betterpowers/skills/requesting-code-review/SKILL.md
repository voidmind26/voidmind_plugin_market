---
name: requesting-code-review
description: Use when you want a second-pass review of a completed task, major feature, diff, or branch before continuing or merging
---

# Requesting Code Review

Dispatch a code reviewer subagent to catch issues before they cascade. The reviewer gets precisely crafted context for evaluation — never your session's history. This keeps the reviewer focused on the work product, not your thought process, and preserves your own context for continued work.

**Core principle:** Review early, review often.

## When to Request Review

**Mandatory:**
- After each task in subagent-driven development
- After completing major feature
- Before merge to main

**Optional but valuable:**
- When stuck (fresh perspective)
- Before refactoring (baseline check)
- After fixing complex bug

## How to Request

**1. Get git SHAs:**

For a single-task or single-phase review, set the review range to the current task or phase only.

```bash
BASE_SHA=$(git rev-parse HEAD~1)
HEAD_SHA=$(git rev-parse HEAD)
```

For multi-phase work, do NOT default to the full branch range. Record the current phase base SHA before implementation starts, then review only the current phase diff:

```bash
PHASE_BASE_SHA=$(git rev-parse HEAD)   # before current phase work starts
PHASE_HEAD_SHA=$(git rev-parse HEAD)   # after current phase work is complete
```

Treat previously accepted work as closed by default. Reopen it only when the current phase modified it, materially depends on it, or an issue outside the current diff makes the current phase invalid.

**2. Dispatch code reviewer subagent:**

Use Task tool with `general-purpose` type and the `sonnet` model, fill template at `code-reviewer.md`

**Placeholders:**
- `{DESCRIPTION}` - Brief summary of what you built in the current task or phase
- `{PLAN_OR_REQUIREMENTS}` - What the current task or phase should do
- `{BASE_SHA}` - Starting commit for the current task or phase
- `{HEAD_SHA}` - Ending commit for the current task or phase

**3. Act on feedback:**
- Fix Critical issues immediately
- Fix Important issues before proceeding
- Note Minor issues for later
- Push back if reviewer is wrong (with reasoning)

## Example

```
[Just completed Phase 2: Add verification function]

You: Let me request code review before proceeding.

PHASE_BASE_SHA=<recorded before Phase 2 started>
PHASE_HEAD_SHA=$(git rev-parse HEAD)

[Dispatch code reviewer subagent]
  DESCRIPTION: Added verifyIndex() and repairIndex() with 4 issue types in Phase 2
  PLAN_OR_REQUIREMENTS: Phase 2 from docs/superpowers/plans/deployment-plan.md
  BASE_SHA: <phase-2-base>
  HEAD_SHA: <phase-2-head>

[Subagent returns]:
  Strengths: Clean architecture, real tests
  Issues:
    Important: Missing progress indicators
    Minor: Magic number (100) for reporting interval
  Out-of-Scope Observations:
    - Phase 1 naming could be clearer, but it is unchanged in this diff
  Assessment: Ready to proceed

You: [Fix progress indicators]
[Continue to Phase 3]
```

## Integration with Workflows

**Subagent-Driven Development:**
- Review after EACH task
- Catch issues before they compound
- Fix before moving to next task

**Executing Plans:**
- Review after each task or at natural checkpoints
- Get feedback, apply, continue

**Ad-Hoc Development:**
- Review before merge
- Review when stuck

## Red Flags

**Never:**
- Skip review because "it's simple"
- Ignore Critical issues
- Proceed with unfixed Important issues
- Argue with valid technical feedback

**If reviewer wrong:**
- Push back with technical reasoning
- Show code/tests that prove it works
- Request clarification

See template at: requesting-code-review/code-reviewer.md
