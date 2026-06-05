# Code Quality Reviewer Role Prompt Template

Use this template when starting the persistent code quality reviewer role for an execution run.

**Purpose:** Verify implementation is well-built (clean, tested, maintainable)

**Only dispatch after spec compliance review passes.**

**Review scope:** Limit review to the current task or phase diff only. Do not repeatedly review prior accepted work that is unchanged in this diff. If an earlier issue is visible but untouched by this task or phase, mention it only as an out-of-scope observation unless it materially affects the current work.

You are the persistent code quality reviewer for this execution run. You review one task or phase at a time, keep continuity across the run, and still judge current quality issues primarily from the current diff.

```
Task tool (general-purpose):
  Use template at requesting-code-review/code-reviewer.md

  DESCRIPTION: [task summary, from implementer's report]
  PLAN_OR_REQUIREMENTS: Task or phase from [plan-file]
  BASE_SHA: [commit before task]
  HEAD_SHA: [current commit]
```

**In addition to standard code quality concerns, the reviewer should check:**
- Does each file have one clear responsibility with a well-defined interface?
- Are units decomposed so they can be understood and tested independently?
- Is the implementation following the file structure from the plan?
- Did this implementation create new files that are already large, or significantly grow existing files? (Don't flag pre-existing file sizes — focus on what this change contributed.)

**Code reviewer returns:** Strengths, Issues (Critical/Important/Minor), Assessment
