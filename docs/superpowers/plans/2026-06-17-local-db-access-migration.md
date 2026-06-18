# local-db-access Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) to implement this plan with fixed-role subagents for implementation, spec review, and code quality review, or superpowers:executing-plans for inline execution. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate `zyb-test-db-plugin` into the private plugin market as `local-db-access` with consistent plugin, MCP, Go module, skill, and marketplace naming.

**Architecture:** Copy the existing plugin into the target market, then perform a focused brand rename while preserving database behavior and security logic. The target market remains the registry source through root `.claude-plugin/marketplace.json`, and the migrated plugin remains self-contained under `plugins/local-db-access/`.

**Tech Stack:** Claude Code plugin marketplace JSON, Go 1.24.2 MCP server, `mcp-go`, MySQL driver, YAML config, zcode/Claude Code skills.

---

## File Structure

**Create:**
- `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/.claude-plugin/plugin.json` — migrated plugin metadata.
- `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/.mcp.json` — MCP service registration using `local-db`.
- `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/README.md` — migrated plugin documentation.
- `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/build.sh` — build helper for `bin/local-db-access-mcp`.
- `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/cmd/test-db-mcp/main.go` — existing Go entrypoint with renamed module imports.
- `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/go.mod` — Go module renamed to `local-db-access`.
- `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/go.sum` — copied dependency checksums.
- `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/internal/**` — copied MCP implementation, with import path rename only.
- `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/skills/**/SKILL.md` — copied skills, with allowed-tools and plugin wording updated.

**Modify:**
- `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/.claude-plugin/marketplace.json` — add `local-db-access` registration.
- `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/README.md` — add plugin directory and registered plugin row.

**Do Not Modify:**
- `/Users/voidmind/GolandProjects/aitools/plugins/zyb-test-db-plugin/**` — source plugin remains unchanged.

---

### Task 1: Copy Plugin Skeleton

- **Intent:** Create a clean target plugin directory containing the source plugin files, excluding generated/local artifacts.
- **Covers spec:** Goals: create `plugins/local-db-access/`; Non-goals: do not modify source plugin.
- **Files:** Create `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/**` from `/Users/voidmind/GolandProjects/aitools/plugins/zyb-test-db-plugin/**`.
- **Granularity rationale:** Copying the plugin is one independently verifiable behavior: target files exist before any rename is applied.

- [ ] **Step 1: Confirm the target directory does not already exist**

Run:

```bash
test ! -e /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access
```

Expected: exit code `0`. If it exits non-zero, stop and inspect the directory instead of overwriting it.

- [ ] **Step 2: Copy source plugin without generated artifacts**

Run:

```bash
rsync -a \
  --exclude '.DS_Store' \
  --exclude 'bin/' \
  /Users/voidmind/GolandProjects/aitools/plugins/zyb-test-db-plugin/ \
  /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/
```

Expected: command exits `0` and creates the target plugin directory.

- [ ] **Step 3: Verify copied source files**

Run:

```bash
cd /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access && \
find . -type f | sort
```

Expected output includes these files and no `./bin/test-db-mcp`:

```text
./.claude-plugin/plugin.json
./.mcp.json
./README.md
./build.sh
./cmd/test-db-mcp/main.go
./go.mod
./go.sum
./internal/common/const.go
./internal/config/config.go
./internal/config/config.yaml
./internal/dto/sql_dto.go
./internal/server/server.go
./internal/tools/db/handler.go
./internal/tools/db/interface.go
./internal/tools/db/manager.go
./internal/tools/db/mysql.go
./internal/tools/db/readonly.go
./internal/tools/db/readonly_test.go
./internal/tools/db/sanitize.go
./skills/init-test-db-config/SKILL.md
./skills/test-db-query/SKILL.md
./skills/test-db-write/SKILL.md
```

- [ ] **Step 4: Verify source plugin is unchanged**

Run:

```bash
git -C /Users/voidmind/GolandProjects/aitools status --short -- plugins/zyb-test-db-plugin
```

Expected: no output.

---

### Task 2: Rename Plugin Runtime Identifiers

- **Intent:** Make the copied plugin build and expose MCP tools under `local-db-access` / `local-db` names.
- **Covers spec:** Design decisions: plugin name, Go module/import rename, MCP service name, binary name, skill allowed-tools rename.
- **Files:** Modify copied `.claude-plugin/plugin.json`, `.mcp.json`, `build.sh`, `go.mod`, `cmd/test-db-mcp/main.go`, `internal/**/*.go`, `skills/**/SKILL.md`.
- **Granularity rationale:** Runtime renaming is one behavior slice because Go build, MCP startup, and skill tool names must agree to be testable.

- [ ] **Step 1: Update plugin metadata**

Replace `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/.claude-plugin/plugin.json` with:

```json
{
  "name": "local-db-access",
  "version": "0.3.3",
  "description": "本地数据库访问插件，提供 MCP 数据库工具与查询/受限写入工作流 skill",
  "author": {
    "name": "voidmind",
    "email": "voidmind@zuoyebang.cc"
  }
}
```

- [ ] **Step 2: Update MCP registration**

Replace `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/.mcp.json` with:

```json
{
  "mcpServers": {
    "local-db": {
      "command": "bash",
      "args": [
        "-lc",
        "BIN=\"$CLAUDE_PLUGIN_ROOT/bin/local-db-access-mcp\"; if [ ! -f \"$BIN\" ]; then cd \"$CLAUDE_PLUGIN_ROOT\" && GOWORK=off go build -o \"$BIN\" ./cmd/test-db-mcp; fi; exec \"$BIN\""
      ]
    }
  }
}
```

- [ ] **Step 3: Update build helper**

Replace `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/build.sh` with:

```bash
#!/bin/bash
# Pre-build the local-db-access-mcp binary so MCP startup is instant.
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

mkdir -p bin

echo "Building local-db-access-mcp..."
GOWORK=off go build -o bin/local-db-access-mcp ./cmd/test-db-mcp
echo "Binary written to bin/local-db-access-mcp"
```

- [ ] **Step 4: Keep build helper executable**

Run:

```bash
chmod +x /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/build.sh
```

Expected: command exits `0`.

- [ ] **Step 5: Rename Go module**

Replace the first line of `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/go.mod`:

```go
module local-db-access
```

Leave the existing `go 1.24.2` and `require` blocks unchanged.

- [ ] **Step 6: Rename Go import paths**

Run:

```bash
cd /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access && \
perl -0pi -e 's#zyb-test-db-plugin/internal/#local-db-access/internal/#g' $(find cmd internal -name '*.go' -type f)
```

Expected: command exits `0`.

- [ ] **Step 7: Rename skill allowed-tools and wording**

Run:

```bash
cd /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access && \
perl -0pi -e 's#mcp__plugin_zyb-test-db-plugin_test-db__#mcp__plugin_local-db-access_local-db__#g; s#zyb-test-db-plugin#local-db-access#g' skills/*/SKILL.md
```

Expected: command exits `0`.

- [ ] **Step 8: Verify runtime identifiers are renamed**

Run:

```bash
rg -n 'zyb-test-db-plugin|mcp__plugin_zyb-test-db-plugin|"test-db"|test-db-mcp|bin/test-db-mcp' /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access
```

Expected: any remaining matches are only intentional references in `README.md` before Task 3 rewrites documentation. There should be no matches in `.claude-plugin/plugin.json`, `.mcp.json`, `build.sh`, `go.mod`, `cmd/`, `internal/`, or `skills/`.

---

### Task 3: Rewrite Plugin Documentation

- **Intent:** Make the migrated plugin documentation match the new plugin name and current plugin-internal config behavior.
- **Covers spec:** Documentation section: update plugin README, remove stale project-level config narrative, update zcode examples.
- **Files:** Modify `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/README.md`.
- **Granularity rationale:** Documentation is separated from runtime renaming so behavior can be reviewed independently from user-facing instructions.

- [ ] **Step 1: Replace plugin README**

Replace `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/README.md` with:

```markdown
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
```

- [ ] **Step 2: Verify README no longer documents old project-level config**

Run:

```bash
rg -n 'zyb-test-db-plugin|\.claude/|default_database_alias|mcp__plugin_zyb-test-db-plugin|test-db-mcp|bin/test-db-mcp' /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access/README.md
```

Expected: no output.

---

### Task 4: Register Plugin in Target Market

- **Intent:** Make `local-db-access` discoverable from the target plugin market and documented in the market README.
- **Covers spec:** Target market `.claude-plugin/marketplace.json` and root `README.md` updates.
- **Files:** Modify `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/.claude-plugin/marketplace.json` and `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/README.md`.
- **Granularity rationale:** Marketplace registration is one externally visible behavior: the market lists and documents the new plugin.

- [ ] **Step 1: Add marketplace entry**

In `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/.claude-plugin/marketplace.json`, add this object to the `plugins` array, keeping alphabetical order by `name` if convenient:

```json
{
  "name": "local-db-access",
  "description": "本地数据库访问插件，提供 MCP 数据库工具与查询/受限写入工作流 skill。",
  "version": "0.3.3",
  "author": {
    "name": "voidmind",
    "email": "voidmind@zuoyebang.cc"
  },
  "source": "./plugins/local-db-access"
}
```

Expected: JSON remains valid.

- [ ] **Step 2: Update market README directory tree**

In `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/README.md`, add this line under `plugins/` entries:

```text
│   ├── local-db-access/           # 本地数据库访问 MCP 插件
```

Expected: directory structure lists `local-db-access` alongside other plugins.

- [ ] **Step 3: Update registered plugin table**

In `/Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/README.md`, add this table row:

```markdown
| [local-db-access](plugins/local-db-access/) | 0.3.3 | 本地数据库访问 MCP 插件，提供查询、受限写入和连接配置初始化能力 |
```

Expected: README registered plugin list includes `local-db-access` exactly once.

- [ ] **Step 4: Validate marketplace JSON**

Run:

```bash
python3 -m json.tool /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/.claude-plugin/marketplace.json >/tmp/local-db-access-marketplace.json
```

Expected: command exits `0`.

- [ ] **Step 5: Verify registration points at the created directory**

Run:

```bash
test -d /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access && \
rg -n '"name": "local-db-access"|"source": "\./plugins/local-db-access"|local-db-access' \
  /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/.claude-plugin/marketplace.json \
  /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/README.md
```

Expected: output shows the marketplace entry and README references.

---

### Task 5: Validate Build, Tests, and Stale Names

- **Intent:** Confirm the migrated plugin builds, tests pass, and no runtime-facing stale names remain.
- **Covers spec:** Testing & verification; error handling around build and tool-name sync.
- **Files:** No intentional source edits unless validation exposes migration mistakes in files already covered by Tasks 2-4.
- **Granularity rationale:** Validation is its own task because it proves the migration rather than introducing new behavior.

- [ ] **Step 1: Run Go tests in the migrated plugin**

Run:

```bash
cd /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access && \
GOWORK=off go test ./...
```

Expected: all packages pass. Existing tests should include readonly SQL guard coverage.

- [ ] **Step 2: Build the renamed binary**

Run:

```bash
cd /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access && \
GOWORK=off go build -o bin/local-db-access-mcp ./cmd/test-db-mcp
```

Expected: command exits `0` and creates `bin/local-db-access-mcp`.

- [ ] **Step 3: Verify build helper uses the same binary**

Run:

```bash
cd /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access && \
./build.sh
```

Expected output includes:

```text
Building local-db-access-mcp...
Binary written to bin/local-db-access-mcp
```

- [ ] **Step 4: Scan runtime files for stale identifiers**

Run:

```bash
rg -n 'zyb-test-db-plugin|mcp__plugin_zyb-test-db-plugin|"test-db"|test-db-mcp|bin/test-db-mcp|default_database_alias|\.claude/zyb-test-db-plugin' \
  /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access \
  /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/.claude-plugin/marketplace.json \
  /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/README.md
```

Expected: no output except any intentional `cmd/test-db-mcp` path references in README or `.mcp.json`. If `cmd/test-db-mcp` path references appear, they are acceptable because the spec explicitly keeps the entrypoint directory.

- [ ] **Step 5: Check target repository diff**

Run:

```bash
git -C /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market status --short && \
git -C /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market diff --stat
```

Expected: changes are limited to:

```text
.claude-plugin/marketplace.json
README.md
docs/superpowers/specs/2026-06-16-local-db-access-migration-design.md
docs/superpowers/plans/2026-06-17-local-db-access-migration.md
plugins/local-db-access/**
```

- [ ] **Step 6: Optional zcode smoke test**

Run only if the environment can launch zcode and the plugin config is safe to use:

```bash
zcode run --plugin-dir /Users/voidmind/Documents/OtherProjects/voidmind_plugin_market/plugins/local-db-access \
  --allowedTools "mcp__plugin_local-db-access_local-db__list_databases" \
  -p "先列出有哪些测试库"
```

Expected: zcode can see the `local-db` MCP service and either lists configured databases or reports a real configuration/connection error from `internal/config/config.yaml`. If skipped, record the skip reason in the final implementation summary.

---

## Self-Review Notes

- **Spec coverage:** Tasks cover copying the plugin, renaming plugin/MCP/Go/skill identifiers, rewriting plugin docs, registering the target market, and validating build/test/stale names.
- **No placeholders:** The plan contains concrete paths, JSON snippets, README content, commands, and expected outcomes.
- **Type consistency:** The selected names are consistent across tasks: plugin `local-db-access`, MCP service `local-db`, binary `local-db-access-mcp`, Go module `local-db-access`.
- **Granularity:** Each task has one verifiable behavior and a bounded file set; no task changes unrelated business logic.
- **Commit policy:** This plan intentionally omits commit steps because the approved spec says not to commit unless the user explicitly requests it.
