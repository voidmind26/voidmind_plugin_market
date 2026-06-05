---
name: code-index-init
description: Use when the user asks to initialize, build, rebuild, or create a code index for the current project.
allowed-tools:
  - mcp__plugin_code-index-plugin_code-index__build_code_index
  - mcp__plugin_code-index-plugin_code-index__get_code_index_status
---

1. Confirm the request is about building or rebuilding the current project's local code index.
2. If it is unclear whether an index already exists, call `get_code_index_status` first.
3. Call `build_code_index` with the current project root and optional deep index paths.
4. Report the index directory plus file, symbol, and chunk counts.
5. After the index is built, ensure `.claude/code-index` is listed in the project's `.gitignore`. If `.gitignore` does not exist, create it. If the entry already exists, skip.
6. If the build fails, surface the tool error directly instead of guessing.
