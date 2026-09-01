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
PLUGIN_VERSION="$(node -e 'const fs = require("fs"); console.log(JSON.parse(fs.readFileSync(".codex-plugin/plugin.json", "utf8")).version)')"
LDFLAGS="-X gateway-platform-plugin/internal/buildinfo.Version=$PLUGIN_VERSION"
GOWORK=off go build -ldflags "$LDFLAGS" -o "$ROOT/bin/gateway-platform-mcp" ./cmd/gateway-platform-mcp
GOWORK=off go build -ldflags "$LDFLAGS" -o "$ROOT/bin/gateway-platform-http" .
