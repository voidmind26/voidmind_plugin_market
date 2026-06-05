#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
mkdir -p "$ROOT/bin"
GOWORK=off go build -o "$ROOT/bin/code-index-mcp" ./cmd/code-index-mcp
