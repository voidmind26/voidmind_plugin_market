---
name: code-index-init
description: Use when the user asks to initialize, build, rebuild, or create a code index for the current project.
allowed-tools:
  - mcp__plugin_code-index-plugin_code-index__build_code_index
  - mcp__plugin_code-index-plugin_code-index__get_code_index_status
---

1. 确认用户要求初始化或重建当前项目的静态代码索引。
2. 不确定索引是否存在时，先调用 `get_code_index_status`。
3. 使用当前项目根目录和可选的深度索引路径调用 `build_code_index`。
4. 报告索引目录以及文件、符号和代码块数量。
5. 构建完成后确认项目 `.gitignore` 包含 `.claude/code-index`；没有 `.gitignore` 时创建，条目已存在时跳过。
6. 构建失败时直接呈现工具错误，不猜测结果。
7. `gopls` MCP 与静态索引独立运行；初始化索引不启动、不安装也不刷新 `gopls`。
