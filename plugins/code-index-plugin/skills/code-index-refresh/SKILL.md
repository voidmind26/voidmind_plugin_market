---
name: code-index-refresh
description: Use when the user asks to refresh, update, or rebuild an existing code index after code changes.
allowed-tools:
  - mcp__plugin_code-index-plugin_code-index__refresh_code_index
  - mcp__plugin_code-index-plugin_code-index__get_code_index_status
  - mcp__plugin_code-index-plugin_code-index__build_code_index
---

1. Call `get_code_index_status` first when you do not know whether an index exists yet.
2. If `ready=false`, tell the user to initialize the index first or call `build_code_index` directly if they asked for a rebuild.
3. Call `refresh_code_index` to rescan the project and update the stored index.
4. Report added, changed, deleted, and unchanged file counts together with the refreshed totals.
5. If refresh fails because index files are missing or invalid, surface the error and recommend rebuilding.
