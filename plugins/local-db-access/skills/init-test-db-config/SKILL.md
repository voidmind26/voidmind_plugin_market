---
name: init-test-db-config
description: This skill should be used when the user asks to "初始化数据库配置", "新增/修改数据库连接", "更新插件内置数据库列表", or wants zcode to write the plugin-internal `internal/config/config.yaml`.
version: 0.2.0
allowed-tools:
  - mcp__plugin_local-db-access_local-db__init_databases
---

# Init Test DB Config

该 skill 用于维护 `local-db-access` 插件内部的数据库连接清单。所有用户共享同一份清单（位于插件目录 `internal/config/config.yaml`），不再使用任何项目级配置。

## 适用范围

在用户要进行以下操作时触发：

- 初始化插件的数据库连接列表
- 新增、修改或下线某条连接
- 调整默认数据库别名（`default_database`）
- 一次性覆写整张连接表

## 工作流

按下面顺序执行：

1. 通过 `list_databases` 读取当前插件内已有的连接和默认库别名。
2. 与用户确认：本次是要 **完整覆盖** 现有清单，还是只是改动其中部分项。
   - 工具底层是覆盖式写入，本 skill 必须先把现有连接补齐再合并用户改动，才能保证不丢连接。
3. 收集每条连接的字段：
   - 连接别名（map key）
   - `type`（默认 `mysql`）
   - `enabled`
   - `host`、`port`、`user`、`password`、`database`
   - `charset`、`timeout`（可选）
4. 让用户明确 `default_database`（可空；为空表示不预设默认库）。
5. 调用 `init_databases`，参数：
   - `overwrite: true`
   - `default_database`
   - `databases`：合并后的完整清单
6. 返回写入路径与连接数，并提醒 **需要重启 MCP 服务（重开 zcode 会话）才能让新配置生效**，因为现有 server 进程仍在使用旧的内存副本。

## 输入约束

- 别名不能为空。
- `default_database` 若提供，必须存在于 `databases` 中。
- 必填字段缺一不可：`host`、`port`、`user`、`password`、`database`。
- 写入是覆盖式的，调用前必须先把要保留的旧连接也带上。

## 输出要求

- 明确写入的配置文件绝对路径。
- 明确写入了多少组连接。
- 明确默认数据库别名（若设置）。
- 提醒重启 MCP 服务。

## 注意事项

- 该配置对所有项目生效，没有项目级覆盖机制。
- 配置中包含明文密码，确保用户清楚该文件不会被自动提交到外部仓库。
- 不写入用户项目目录，也不修改任何 `CLAUDE.md`。
