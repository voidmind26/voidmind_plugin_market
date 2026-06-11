#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
FRONTEND="$ROOT/frontend"
EMBED_DIST="$ROOT/server/router/frontend_dist"

cd "$FRONTEND"
if [ ! -d node_modules ] || [ ! -x node_modules/.bin/vite ]; then
  npm install >/dev/null
fi
if ! npm run build >/dev/null; then
  rm -rf node_modules package-lock.json
  npm install >/dev/null
  npm run build >/dev/null
fi

rm -rf "$EMBED_DIST"
mkdir -p "$EMBED_DIST"
cp -R "$FRONTEND/dist/." "$EMBED_DIST/"

cd "$ROOT"
GOWORK=off go test ./server/...
mkdir -p "$ROOT/bin"
GOWORK=off go build -o "$ROOT/bin/gateway-platform-mcp" ./cmd/gateway-platform-mcp
