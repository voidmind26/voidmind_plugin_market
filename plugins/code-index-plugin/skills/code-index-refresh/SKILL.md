---
name: code-index-refresh
description: Use when the user asks to refresh, update, or rebuild an existing code index after code changes.
allowed-tools:
  - mcp__plugin_code-index-plugin_code-index__refresh_code_index
  - mcp__plugin_code-index-plugin_code-index__get_code_index_status
  - mcp__plugin_code-index-plugin_code-index__build_code_index
---

1. 不确定索引是否存在时，先调用 `get_code_index_status`。
2. `ready=false` 时说明需要初始化；用户要求重建时可直接调用 `build_code_index`。
3. 调用 `refresh_code_index` 重新扫描项目并更新静态索引。
4. 报告新增、变更、删除和未变文件数量，以及刷新后的总量。
5. 索引文件缺失或损坏导致刷新失败时，呈现原始错误并建议重建。
6. 刷新只影响静态索引，不重启、不安装也不刷新路由层管理的 `gopls` 子会话。
