#!/usr/bin/env bash
# Test: test-driven-development skill follows E2E-first workflow
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/test-helpers.sh"

echo "=== Test: test-driven-development skill (E2E-first) ==="
echo ""

# Test 1: 默认从大粒度测试起步
# 这里不要求 Claude 真写代码，只验证它如何描述 TDD 起手式

echo "Test 1: E2E-first default..."
output=$(run_claude "In the test-driven-development skill, for a normal feature request in a workflow-heavy codebase, what kind of failing test should come first? Answer briefly." 30)

if assert_contains "$output" "E2E\|end-to-end\|integration" "Mentions E2E or integration first"; then
    :
else
    exit 1
fi

echo ""

# Test 2: 第一条测试应带复杂输入

echo "Test 2: Complex input requirement..."
output=$(run_claude "According to the test-driven-development skill, should the first failing test use a trivial happy-path input or a relatively complex input?" 30)

if assert_contains "$output" "complex\|complexity\|realistic\|representative" "Requires relatively complex input"; then
    :
else
    exit 1
fi

if assert_not_contains "$output" "trivial happy-path only\|only the simplest" "Does not prefer trivial happy-path only"; then
    :
else
    exit 1
fi

echo ""

# Test 3: 不默认补齐小粒度测试

echo "Test 3: No automatic unit-test backfill..."
output=$(run_claude "If the large-granularity E2E test passes cleanly, does the test-driven-development skill require adding unit tests for every new function anyway?" 30)

if assert_contains "$output" "no\|not required\|does not require\|only if" "Does not require blanket unit-test backfill"; then
    :
else
    exit 1
fi

echo ""

# Test 4: 只有失败或异常时才下钻

echo "Test 4: Escalate down only on failure or anomalies..."
output=$(run_claude "When does the test-driven-development skill allow dropping down from E2E tests to smaller unit or module tests?" 30)

if assert_contains "$output" "fail\|failure\|flaky\|timing\|performance\|stability" "Mentions failure or anomalies as trigger"; then
    :
else
    exit 1
fi

if assert_not_contains "$output" "always start with unit tests\|normally begin with unit tests" "Does not fall back to unit tests by default"; then
    :
else
    exit 1
fi

echo ""

# Test 5: 天然不适合 E2E 的例外路径

echo "Test 5: Exception path for non-E2E tasks..."
output=$(run_claude "If the task is a pure utility function with no natural end-to-end flow, what does the test-driven-development skill say to do first?" 30)

if assert_contains "$output" "unit\|module\|smaller-scope\|local test" "Allows smaller-scope tests for non-E2E tasks"; then
    :
else
    exit 1
fi

if assert_contains "$output" "test-first\|failing test first" "Keeps test-first even in exception path"; then
    :
else
    exit 1
fi

echo ""
echo "=== All test-driven-development E2E-first tests passed ==="