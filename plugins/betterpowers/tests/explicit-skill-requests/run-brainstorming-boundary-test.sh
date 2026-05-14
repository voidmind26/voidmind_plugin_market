#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
PROMPTS_DIR="$SCRIPT_DIR/prompts"

TIMESTAMP=$(date +%s)
OUTPUT_DIR="/tmp/superpowers-tests/${TIMESTAMP}/brainstorming-boundaries"
PROJECT_DIR="$OUTPUT_DIR/project"
LOG_CLEAR="$OUTPUT_DIR/clear-norms.json"

mkdir -p "$PROJECT_DIR/docs/superpowers/specs" "$OUTPUT_DIR"
cd "$PROJECT_DIR"

PROMPT_CLEAR=$(cat "$PROMPTS_DIR/brainstorming-clear-norms.txt")

timeout 300 claude -p "$PROMPT_CLEAR" \
  --plugin-dir "$PLUGIN_DIR" \
  --dangerously-skip-permissions \
  --max-turns 4 \
  --output-format stream-json \
  > "$LOG_CLEAR" 2>&1 || true

if ! grep -q 'docs/superpowers/specs/' "$LOG_CLEAR"; then
  echo "FAIL: clear-norms scenario did not mention spec output path"
  exit 1
fi

if grep -qiE 'what (should|do you)|which option|can you clarify|want me to' "$LOG_CLEAR"; then
  echo "FAIL: clear-norms scenario asked an unnecessary clarification question"
  exit 1
fi

echo "PASS: clear-norms scenario wrote spec without unnecessary questions"
