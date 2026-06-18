#!/bin/bash
# Pre-build the local-db-access-mcp binary so MCP startup is instant.
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

mkdir -p bin

echo "Building local-db-access-mcp..."
GOWORK=off go build -o bin/local-db-access-mcp ./cmd/test-db-mcp
echo "Binary written to bin/local-db-access-mcp"
