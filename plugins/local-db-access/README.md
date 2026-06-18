# local-db-access

本地数据库访问插件，提供一组可被 zcode / Claude Code 直接接入的 MCP 数据库工具，以及三个 skill：数据库查询流程、测试库受限写入流程、插件内置数据库连接配置初始化流程。

## 插件目标

该插件面向受控测试环境或本地可访问数据库，封装数据库工具族，支持：

- 列出可用数据库连接
- 查看数据库整体信息
- 查看表结构
- 查看表数据样本
- 执行只读 SQL
- 执行受限写入 SQL（INSERT / UPDATE / CREATE / ALTER）
- 检查数据库连接健康状态
- 初始化或覆盖插件内部数据库连接配置

## 目录结构

```text
local-db-access/
├── .claude-plugin/plugin.json
├── .mcp.json
├── README.md
├── build.sh
├── cmd/test-db-mcp/main.go
├── go.mod
├── go.sum
├── bin/                  (编译产物，已忽略)
│   └── local-db-access-mcp
├── internal/
│   ├── common/
│   ├── config/
│   ├── dto/
│   ├── server/
│   └── tools/db/
└── skills/
    ├── test-db-query/SKILL.md
    ├── test-db-write/SKILL.md
    └── init-test-db-config/SKILL.md
```

## MCP 接入方式

插件通过根目录下的 `.mcp.json` 注册本地 `stdio` MCP 服务。

- 服务名：`local-db`
- 启动方式：优先使用预编译的 `bin/local-db-access-mcp` 二进制文件；若不存在则自动编译后运行

首次启动时会有一次编译开销，后续启动直接运行二进制文件，无需重复编译。

你也可以手动预编译：

```bash
./build.sh
# 或
GOWORK=off go build -o bin/local-db-access-mcp ./cmd/test-db-mcp
```

启动脚本显式设置 `GOWORK=off`，避免外层 Go workspace 干扰插件独立运行。

## 提供的 MCP 工具

### 1. `list_databases`
列出当前运行时可用的数据库连接。

### 2. `get_database_info`
查看指定数据库的基础信息，包括版本、当前库、表数量、表摘要与库大小信息。

### 3. `describe_table`
查看指定表的列结构、主键/唯一索引与建表语句。

### 4. `get_table_data`
按 limit 获取指定表的数据样本。

### 5. `execute_query`
执行只读 SQL 查询，仅允许：

- `SELECT`
- `SHOW`
- `DESC`
- `DESCRIBE`
- `EXPLAIN`

插件会拦截写操作与危险语句。

### 6. `execute_write_query`
执行受限写入 SQL，仅允许：

- `INSERT`
- `UPDATE`
- `CREATE`
- `ALTER`

插件会拒绝：

- `DELETE`
- `DROP`
- `TRUNCATE`
- 多语句拼接

### 7. `health_check`
检查单个数据库或全部数据库连接是否健康。

### 8. `init_databases`
初始化或覆盖插件内部数据库连接配置，写入 `internal/config/config.yaml`。

## 插件内部配置

数据库连接集合来源只有插件内部配置文件：

```text
internal/config/config.yaml
```

示例：

```yaml
default_database: project_main
databases:
  project_main:
    type: mysql
    enabled: true
    host: 10.0.0.1
    port: 3306
    user: demo_user
    password: demo_pass
    database: project_main
    charset: utf8mb4
    timeout: 5
  project_shadow:
    type: mysql
    enabled: true
    host: 10.0.0.2
    port: 3306
    user: shadow_user
    password: demo_pass
    database: project_shadow
    charset: utf8mb4
    timeout: 5
```

## 数据库选择规则

默认数据库选择按下面顺序处理：

1. 用户本次明确指定的 `database_name`
2. `list_databases` 返回的 `default_database`
3. 如果以上都没有，则先 `list_databases`，再让用户确认目标连接

## Skill 说明

### 1. `test-db-query`
指导 zcode 按安全顺序查询数据库：

1. 未指定库时先列库
2. 优先看结构和样本
3. 最后才执行只读 SQL

### 2. `test-db-write`
指导 zcode 安全执行受限写入：

1. 未指定库时先列库
2. 必要时先看结构和样本
3. 仅允许 `INSERT` / `UPDATE` / `CREATE` / `ALTER`
4. 明确拒绝 `DELETE` / `DROP` / `TRUNCATE`
5. 调用 `execute_write_query`

### 3. `init-test-db-config`
指导 zcode 初始化或覆盖插件内部数据库连接配置：

1. 读取当前连接清单
2. 确认完整覆盖或合并局部改动
3. 收集多组连接信息
4. 指定默认数据库
5. 调用 `init_databases`

## 本地验证

在插件目录内执行：

```bash
GOWORK=off go build ./...
GOWORK=off go test ./...
GOWORK=off go run ./cmd/test-db-mcp
```

## 在 zcode 中测试插件

可以通过插件目录直接测试：

```bash
zcode run --plugin-dir /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access
```

非交互最小验证示例：

```bash
zcode run --plugin-dir /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access \
  --allowedTools "mcp__plugin_local-db-access_local-db__list_databases" \
  -p "先列出有哪些测试库"
```

## 安全边界

- 本插件仅面向受控测试环境或本地可访问数据库
- 读操作仅支持只读查询；写操作仅支持 `INSERT` / `UPDATE` / `CREATE` / `ALTER`
- 明确禁用 `DELETE` / `DROP` / `TRUNCATE` 与多语句拼接
- 数据库连接配置位于插件内部 `internal/config/config.yaml`
- 配置中可能包含明文密码，提交前必须确认是否允许进入目标市场仓库
