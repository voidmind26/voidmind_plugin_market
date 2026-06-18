---
name: test-db-write
description: This skill should be used when the user asks to "修改测试库数据", "往测试库插入数据", "新建测试表", "改测试库表结构", "执行 insert/update/create/alter SQL", or wants zcode to safely write to test databases without allowing delete statements.
version: 0.1.0
allowed-tools:
  - mcp__plugin_local-db-access_local-db__list_databases
  - mcp__plugin_local-db-access_local-db__describe_table
  - mcp__plugin_local-db-access_local-db__get_table_data
  - mcp__plugin_local-db-access_local-db__execute_query
  - mcp__plugin_local-db-access_local-db__execute_write_query
  - mcp__plugin_local-db-access_local-db__health_check
---

# Test DB Write

该 skill 用于指导 zcode 在测试环境数据库中安全执行受限写入 SQL。

## 适用范围

在用户要进行以下操作时触发：

- 修改测试库数据
- 往测试库插入数据
- 新建测试表
- 修改测试库表结构
- 执行 `INSERT` / `UPDATE` / `CREATE` / `ALTER` SQL
- 在测试数据库中准备或修正测试数据

## 数据库选择规则

数据库连接清单只有一个来源：插件内部 `internal/config/config.yaml`（对话中展现为 `list_databases` 返回的 `databases`）。

选择目标连接时按以下顺序：

1. 用户本次对话中明确指定的别名。
2. 用户在对话中使用的"模型别名"——与 `list_databases` 返回的 `alias` 做匹配；若匹配唯一就直接用，否则必须向用户确认。
3. `list_databases` 返回的 `default_database`（若非空）。
4. 以上都不满足：先调用 `list_databases` 展示清单，让用户确认。

## 工作流

按下面顺序执行，不要跳步猜测数据库：

1. 识别是否是测试环境数据库写入需求。
2. 如果用户没有明确别名，先调用 `list_databases` 获取连接清单 + 各库下的表名，再与用户确认目标连接。
3. 如有必要，先调用 `describe_table`、`get_table_data` 或 `execute_query` 了解当前结构和样本数据。
4. 仅当 SQL 属于 `INSERT`、`UPDATE`、`CREATE`、`ALTER` 时，才调用 `execute_write_query`。
5. 如果用户请求 `DELETE`、`DROP`、`TRUNCATE`，明确拒绝并说明插件禁用该类语句。
6. 返回结果时说明使用了哪个 `database_name`、执行的是哪类语句、影响了多少行。

## 工具使用原则

### 先列库，再选库

当用户没有明确别名时，不要猜测。先调用 `list_databases`，其返回里已经包含每个连接下的全部表名，通常一次调用就足以帮助用户选型。

### 先结构/样本，再写入

- 想确认表结构 → `describe_table`
- 想确认样本数据 → `get_table_data`
- 想复杂筛选排查 → `execute_query`
- 只有确认写入 SQL 合法后，才调用 `execute_write_query`

### 保持受限写入边界

`execute_write_query` 仅允许：

- `INSERT`
- `UPDATE`
- `CREATE`
- `ALTER`

不得尝试以下语句：

- `DELETE`
- `DROP`
- `TRUNCATE`
- 多语句拼接

## 结果表达要求

- 明确说明当前操作面向测试环境数据库。
- 在返回结果前说明使用了哪个 `database_name`，以及该别名是用户指定、默认库还是匹配得到的。
- 说明执行的是哪类语句，影响行数是多少。
- 如果是被插件策略拒绝，如实说明是语句类型不允许，不臆测业务原因。

## 注意事项

- 插件内置的是测试环境连接，不代表生产环境。
- 所有工具都以别名（而非真实 MySQL database 名）作为 `database_name` 入参。
- 插件仅开放受限写入，不代表允许任意 SQL。
