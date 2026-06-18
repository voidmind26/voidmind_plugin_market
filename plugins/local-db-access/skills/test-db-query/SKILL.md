---
name: test-db-query
description: This skill should be used when the user asks to "查询测试库", "查看测试数据库", "看表结构", "查表数据", "执行只读 SQL", "先列出有哪些测试库", or wants zcode to inspect test databases safely before querying.
version: 0.2.0
allowed-tools:
  - mcp__plugin_local-db-access_local-db__list_databases
  - mcp__plugin_local-db-access_local-db__get_database_info
  - mcp__plugin_local-db-access_local-db__describe_table
  - mcp__plugin_local-db-access_local-db__get_table_data
  - mcp__plugin_local-db-access_local-db__execute_query
  - mcp__plugin_local-db-access_local-db__health_check
  - mcp__plugin_local-db-access_local-db__init_databases
---

# Test DB Query

该 skill 用于指导 zcode 在测试环境数据库场景下正确使用 `local-db-access` 提供的 MCP 工具族。

## 适用范围

在用户要进行以下操作时触发：

- 查询测试环境数据库
- 查看有哪些测试库与表
- 查看某张表的结构或数据样本
- 执行只读 SQL 排查数据问题
- 新增、修改或覆盖插件内部的数据库连接配置

## 数据库选择规则

数据库连接清单只有一个来源：插件内部 `internal/config/config.yaml`（对话中展现为 `list_databases` 返回的 `databases`）。

选择目标连接时按以下顺序：

1. 用户本次对话中明确指定的别名。
2. 用户在对话中使用的"模型别名"——与 `list_databases` 返回的 `alias` 做匹配；若匹配唯一就直接用，否则必须向用户确认。
3. `list_databases` 返回的 `default_database`（若非空）。
4. 以上都不满足：先调用 `list_databases` 展示清单，让用户确认。

## 工作流

按下面顺序执行，不要跳步猜测数据库：

1. 识别是否是测试环境数据库查询需求。
2. 如果用户没有明确别名，先调用 `list_databases` 获取连接清单 + 各库下的表名，再与用户确认目标连接。
3. 选定 `database_name` 后再执行针对具体库的工具。
4. 如果目标是理解数据库全貌，调用 `get_database_info`。
5. 如果目标是理解表结构，调用 `describe_table`。
6. 如果目标是查看样本数据，调用 `get_table_data`。
7. 只有在需要灵活条件排查时，才调用 `execute_query`。
8. 如果用户要求检查连接问题，调用 `health_check`。
9. 如果用户要维护连接清单本身，转到 `init-test-db-config` skill。

## 工具使用原则

### 先列库，再选库

当用户没有明确别名时，不要猜测。先调用 `list_databases`，其返回里已经包含每个连接下的全部表名，通常一次调用就足以帮助用户选型。

### 先结构/样本，再裸 SQL

- 想"看看这个库长啥样" → `get_database_info` 或 `describe_table`。
- 想"看看数据大致长啥样" → `get_table_data`。
- 只有筛选条件复杂或用户明确要求，才用 `execute_query`。

### 保持只读边界

`execute_query` 仅用于只读 SQL（SELECT/SHOW/DESCRIBE/EXPLAIN）。不得尝试任何写/DDL/清理操作。

## 结果表达要求

- 明确说明当前操作面向测试环境数据库。
- 在返回数据前说明使用了哪个 `database_name`，以及该别名是用户指定、默认库还是匹配得到的。
- 如果没有数据，如实说明为空，不臆测业务原因。
- 如果用户未指定且没有唯一匹配，明确告知已先列库。

## 注意事项

- 插件内置的是测试环境连接，不代表生产环境。
- 所有工具都以别名（而非真实 MySQL database 名）作为 `database_name` 入参。
- 配置发生改动后需要重启 MCP 服务才能让新连接在工具调用中生效。
