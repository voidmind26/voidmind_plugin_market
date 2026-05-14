#!/usr/bin/env bash
# Test: Do worktree-related skills resolve the target repo from task context first?
#
# RED: current skills do not document task-repo priority or ask-user fallback.
# GREEN: updated skills should prefer the repo named by the current task and ask the user when repo identity is still ambiguous.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/test-helpers.sh"

failures=0

echo "=== Worktree Target Repo Resolution Test ==="
echo ""

echo "--- Test 1: using-git-worktrees prefers task repo over parent directory ---"
output=$(run_claude "According to the using-git-worktrees skill, if the current task explicitly targets the child repo uos-game-web but the current working directory is a parent folder containing multiple repos, which repo should be treated as the target for worktree setup? Answer briefly." 30)
echo "$output"
echo ""
assert_contains "$output" "uos-game-web" "Names the task repo as target" || failures=$((failures + 1))
assert_contains "$output" "task" "Explains task context takes priority" || failures=$((failures + 1))
echo ""

echo "--- Test 2: using-git-worktrees asks the user when repo is still ambiguous ---"
output=$(run_claude "According to the using-git-worktrees skill, if task context, current directory, and child-repo discovery still cannot uniquely determine the target repo, what should the agent do next? Answer briefly." 30)
echo "$output"
echo ""
assert_contains "$output" "ask" "Requires asking the user" || failures=$((failures + 1))
assert_contains "$output" "user" "Directs ambiguity back to the user" || failures=$((failures + 1))
echo ""

echo "--- Test 3: finishing-a-development-branch uses the same target repo rule ---"
output=$(run_claude "According to the finishing-a-development-branch skill, if the implementation task is for child repo uos-game-web but the open folder is a parent directory, which repo should Step 2 environment detection and cleanup use? Answer briefly." 30)
echo "$output"
echo ""
assert_contains "$output" "uos-game-web" "Finishing skill targets the task repo" || failures=$((failures + 1))
assert_contains "$output" "task" "Finishing skill preserves task-repo priority" || failures=$((failures + 1))
echo ""

if [ "$failures" -eq 0 ]; then
    echo "=== ALL TESTS PASSED ==="
else
    echo "=== $failures TEST(S) FAILED ==="
    exit 1
fi
