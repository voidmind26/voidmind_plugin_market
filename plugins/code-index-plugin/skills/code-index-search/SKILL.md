---
name: code-index-search
description: Use when the user asks to locate code structure or implementation details and a local code index can narrow the search scope before reading files.
allowed-tools:
  - mcp__plugin_code-index-plugin_code-index__search_code_index
  - mcp__plugin_code-index-plugin_code-index__get_code_index_status
  - mcp__plugin_code-index-plugin_code-index__build_code_index
---

1. Extract concise search terms from the user request.
2. Call `get_code_index_status` to check whether a code index already exists for the project.
3. If `ready=false` or the index does not exist, call `build_code_index` first to create the index.
4. Call `search_code_index` with the query.
5. Prefer `prefer_deep_hits=true` when the user is asking for concrete implementation points such as functions or handlers.
6. Return the top hits with kind, path, line range, summary, and score reason.
7. Only read source files after the index has narrowed the scope.
