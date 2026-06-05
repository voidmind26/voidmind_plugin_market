---
name: using-git-worktrees
description: Use when starting work that should be isolated from the current branch, when the user asks for a worktree, or before executing a plan in a separate workspace
---

# Using Git Worktrees

## Overview

Ensure work happens in an isolated workspace. Prefer your platform's native worktree tools. Fall back to manual git worktrees only when no native tool is available.

**Core principle:** Resolve the target repository first. Confirm the base branch and output branch before creation. Then detect existing isolation. Then use native tools. Then fall back to git. Never fight the harness.

**Announce at start:** "I'm using the using-git-worktrees skill to set up an isolated workspace."

## Step 0: Resolve Target Repository and Detect Existing Isolation

**Before creating anything, resolve which repository this task is actually about.**

Use this priority order:
1. If the current task already names or clearly points to a repository, use that repository.
2. Otherwise, if the current directory is already inside a git repository, use that repository.
3. Otherwise, look for git repositories under the current directory.
4. If that discovery finds exactly one candidate, use it.
5. If the repository is still ambiguous or no candidate exists, ask the user which repository to use. Do not guess.

**All later git commands in this skill should run against that target repository, even if your session started from a parent directory.**

Once the target repository is known, check whether it is already in an isolated workspace.

```bash
GIT_DIR=$(cd "$(git -C "$TARGET_REPO" rev-parse --git-dir)" 2>/dev/null && pwd -P)
GIT_COMMON=$(cd "$(git -C "$TARGET_REPO" rev-parse --git-common-dir)" 2>/dev/null && pwd -P)
BRANCH=$(git -C "$TARGET_REPO" branch --show-current)
```

**Submodule guard:** `GIT_DIR != GIT_COMMON` is also true inside git submodules. Before concluding "already in a worktree," verify you are not in a submodule:

```bash
# If this returns a path, you're in a submodule, not a worktree — treat as normal repo
git -C "$TARGET_REPO" rev-parse --show-superproject-working-tree 2>/dev/null
```

**If `GIT_DIR != GIT_COMMON` (and not a submodule):** The target repository is already in a linked worktree. Skip to Step 3 (Project Setup). Do NOT create another worktree.

Report with branch state:
- On a branch: "Already in isolated workspace at `<path>` on branch `<name>`."
- Detached HEAD: "Already in isolated workspace at `<path>` (detached HEAD, externally managed). Branch creation needed at finish time."

**If `GIT_DIR == GIT_COMMON` (or in a submodule):** The target repository is in a normal repo checkout.

Before creating anything, resolve two branch values explicitly:
- **Base branch:** which branch the worktree should be created from
- **Output branch:** which branch the implementation will live on inside the worktree

These are branch roles, not a rule that the names must differ. They may be the same branch name.

If your instructions or repository context already make both values clear, state them before creation.

If either value is unclear, ask before creating the worktree. Do not guess.

Use a direct prompt like:

> "Before I create the worktree: which branch should I base it on, and what output branch name should I use for the work?"

Has the user already indicated their worktree preference in your instructions? If not, ask for consent before creating a worktree:

> "Would you like me to set up an isolated worktree? It protects your current branch from changes."

Honor any existing declared preference without asking. If the user declines consent, work in place and skip to Step 3.

## Step 1: Create Isolated Workspace

**You have two mechanisms. Try them in this order.**

### 1a. Native Worktree Tools (preferred)

The user has asked for an isolated workspace (Step 0 consent), and the base branch plus output branch are already confirmed. Do you already have a way to create a worktree? It might be a tool with a name like `EnterWorktree`, `WorktreeCreate`, a `/worktree` command, or a `--worktree` flag. If you do, use it and skip to Step 3.

Native tools handle directory placement, branch creation, and cleanup automatically. Using `git worktree add` when you have a native tool creates phantom state your harness can't see or manage.

Only proceed to Step 1b if you have no native worktree tool available.

### 1b. Git Worktree Fallback

**Only use this if Step 1a does not apply** — you have no native worktree tool available. Create a worktree manually using git.

#### Directory Selection

Follow this priority order. Explicit user preference always beats observed filesystem state.

1. **Check your instructions for a declared worktree directory preference.** If the user has already specified one, use it without asking.

2. **Check for an existing project-local worktree directory in the target repository:**
   ```bash
   ls -d "$TARGET_REPO/.worktrees" 2>/dev/null     # Preferred (hidden)
   ls -d "$TARGET_REPO/worktrees" 2>/dev/null      # Alternative
   ```
   If found, use it. If both exist, `.worktrees` wins.

3. **Check for an existing global directory:**
   ```bash
   project=$(basename "$(git -C "$TARGET_REPO" rev-parse --show-toplevel)")
   ls -d ~/.config/superpowers/worktrees/$project 2>/dev/null
   ```
   If found, use it (backward compatibility with legacy global path).

4. **If there is no other guidance available**, default to `.worktrees/` at the project root.

#### Safety Verification (project-local directories only)

**MUST verify directory is ignored before creating worktree:**

```bash
git -C "$TARGET_REPO" check-ignore -q .worktrees 2>/dev/null || git -C "$TARGET_REPO" check-ignore -q worktrees 2>/dev/null
```

**If NOT ignored:** Add to .gitignore, commit the change, then proceed.

**Why critical:** Prevents accidentally committing worktree contents to repository.

Global directories (`~/.config/superpowers/worktrees/`) need no verification.

#### Create the Worktree

Before creating the worktree, state the branch plan explicitly:
- `Base branch: <base-branch>`
- `Output branch: <output-branch>`

`<output-branch>` is the branch that will contain the implementation and will later be merged, pushed, or kept. It may be the same as `<base-branch>`.

```bash
project=$(basename "$(git -C "$TARGET_REPO" rev-parse --show-toplevel)")

# Determine path based on chosen location
# For project-local: path="$LOCATION/$OUTPUT_BRANCH"
# For global: path="~/.config/superpowers/worktrees/$project/$OUTPUT_BRANCH"

if [ "$OUTPUT_BRANCH" = "$BASE_BRANCH" ]; then
  git -C "$TARGET_REPO" worktree add "$path" "$BASE_BRANCH"
else
  git -C "$TARGET_REPO" worktree add "$path" -b "$OUTPUT_BRANCH" "$BASE_BRANCH"
fi
cd "$path"
```

**Sandbox fallback:** If `git worktree add` fails with a permission error (sandbox denial), tell the user the sandbox blocked worktree creation and you're working in the current directory instead. Then run setup and baseline tests in place.

## Step 2: Finishing Work in the Worktree

When implementation is done, default to a git-native handoff:

1. Commit the finished changes inside the worktree branch.
2. Return to the main checkout.
3. Merge or cherry-pick the worktree commit into the destination branch.
4. Only then clean up the worktree.

**Do not manually copy edited files from the worktree back into the main checkout** as your default completion path. That loses provenance, makes rollback harder, and increases the chance of missing companion files.

**Use manual copy only for deliberate recovery** when the git-native path is no longer viable and the user accepts that trade-off.

## Step 3: Project Setup

Auto-detect and run appropriate setup:

```bash
# Node.js
if [ -f package.json ]; then npm install; fi

# Rust
if [ -f Cargo.toml ]; then cargo build; fi

# Python
if [ -f requirements.txt ]; then pip install -r requirements.txt; fi
if [ -f pyproject.toml ]; then poetry install; fi

# Go
if [ -f go.mod ]; then go mod download; fi
```

## Step 4: Verify Clean Baseline

Run tests to ensure workspace starts clean:

```bash
# Use project-appropriate command
npm test / cargo test / pytest / go test ./...
```

**If tests fail:** Report failures, ask whether to proceed or investigate.

**If tests pass:** Report ready.

### Report

```
Worktree ready at <full-path>
Base branch: <base-branch>
Output branch: <output-branch>
Tests passing (<N> tests, 0 failures)
Ready to implement <feature-name>
```

## Quick Reference

| Situation | Action |
|-----------|--------|
| Task already identifies repo | Use that repo first |
| Repo still ambiguous after context + discovery | Ask user, don't guess |
| Base branch or output branch unclear | Ask user before creation |
| Already in linked worktree | Skip creation (Step 0) |
| In a submodule | Treat as normal repo (Step 0 guard) |
| Native worktree tool available | Use it after branch values are confirmed (Step 1a) |
| No native tool | Git worktree fallback (Step 1b) |
| `.worktrees/` exists | Use it (verify ignored) |
| `worktrees/` exists | Use it (verify ignored) |
| Both exist | Use `.worktrees/` |
| Neither exists | Check instruction file, then default `.worktrees/` |
| Global path exists | Use it (backward compat) |
| Directory not ignored | Add to .gitignore + commit |
| Permission error on create | Sandbox fallback, work in place |
| Tests fail during baseline | Report failures + ask |
| No package.json/Cargo.toml | Skip dependency install |

## Common Mistakes

### Fighting the harness

- **Problem:** Using `git worktree add` when the platform already provides isolation
- **Fix:** Step 0 detects existing isolation. Step 1a defers to native tools.

### Skipping detection

- **Problem:** Creating a nested worktree inside an existing one or for the wrong repository
- **Fix:** Always run Step 0 before creating anything

### Skipping ignore verification

- **Problem:** Worktree contents get tracked, pollute git status
- **Fix:** Always use `git check-ignore` before creating project-local worktree

### Assuming repository from the open folder

- **Problem:** Parent directories can contain multiple repositories, so the open folder may not identify the task repo
- **Fix:** Follow priority: task repo > current repo > unique discovered child repo > ask user

### Assuming directory location

- **Problem:** Creates inconsistency, violates project conventions
- **Fix:** Follow priority: existing > global legacy > instruction file > default

### Skipping branch confirmation

- **Problem:** Worktree gets created from the wrong base or on an unclear implementation branch
- **Fix:** Confirm both base branch and output branch before creation, and report both after creation

### Proceeding with failing tests

- **Problem:** Can't distinguish new bugs from pre-existing issues
- **Fix:** Report failures, get explicit permission to proceed

## Red Flags

**Never:**
- Create a worktree when Step 0 detects existing isolation
- Guess which repository the task means when context is still ambiguous
- Guess the base branch or output branch when either is unclear
- Use `git worktree add` when you have a native worktree tool (e.g., `EnterWorktree`). This is the #1 mistake — if you have it, use it.
- Skip Step 1a by jumping straight to Step 1b's git commands
- Create worktree without verifying it's ignored (project-local)
- Skip baseline test verification
- Proceed with failing tests without asking

**Always:**
- Resolve the task's target repository before running git commands
- Ask the user if the target repository is still ambiguous
- Confirm the base branch and output branch before creating the worktree
- Run Step 0 detection first
- Prefer native tools over git fallback
- Follow directory priority: existing > global legacy > instruction file > default
- Verify directory is ignored for project-local
- Auto-detect and run project setup
- Report the worktree path plus both branch values after creation
- Verify clean test baseline
