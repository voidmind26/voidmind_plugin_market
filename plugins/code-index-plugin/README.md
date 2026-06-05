# code-index-plugin

为当前项目构建本地代码索引，并通过 Claude Code MCP 工具提供构建、刷新、搜索和状态查询能力。

## 插件目标

该插件面向 Claude Code 的代码定位场景，提供一套本地索引能力：

- 构建当前项目的文件级索引
- 为 Go 文件提取结构化符号索引
- 为代码和文档提取启发式 chunk 索引
- 在代码变更后刷新本地索引
- 在搜索实现位置前先查询本地索引，缩小文件读取范围

第一版目标是：

- 加快代码定位速度
- 减少无效整文件读取
- 降低搜索链路中的 token 消耗

## 目录结构

```text
code-index-plugin/
├── .claude-plugin/plugin.json
├── .mcp.json
├── README.md
├── build.sh
├── cmd/code-index-mcp/main.go
├── go.mod
├── go.sum
├── internal/
│   ├── index/
│   │   ├── config/
│   │   ├── extractor/
│   │   ├── manifest/
│   │   ├── model/
│   │   ├── query/
│   │   ├── scanner/
│   │   ├── service/
│   │   └── storage/
│   ├── server/
│   └── tools/index/
└── skills/
    ├── code-index-init/
    ├── code-index-refresh/
    └── code-index-search/
```

## MCP 工具

插件通过 `.mcp.json` 注册本地 `stdio` MCP 服务，提供 4 个工具：

### 1. `build_code_index`
构建当前项目的本地索引。

输入：
- `project_root`（可选）
- `deep_index_paths`（可选）

输出：
- `project_root`
- `index_dir`
- `file_count`
- `symbol_count`
- `chunk_count`

### 2. `refresh_code_index`
刷新已存在的本地索引。

输入：
- `project_root`（可选）
- `deep_index_paths`（可选）

输出：
- `added_count`
- `changed_count`
- `deleted_count`
- `unchanged_count`
- 刷新后的索引统计

### 3. `search_code_index`
搜索当前项目索引。

输入：
- `query`
- `project_root`（可选）
- `path_prefix`（可选）
- `prefer_deep_hits`（可选）
- `limit`（可选）

输出：
- `result_count`
- `results[]`
  - `kind`
  - `path`
  - `title`
  - `start_line`
  - `end_line`
  - `summary`
  - `score`
  - `score_reason`

### 4. `get_code_index_status`
查看索引状态。

输入：
- `project_root`（可选）

输出：
- `ready`
- `project_root`
- `index_dir`
- `file_count`
- `symbol_count`
- `chunk_count`
- `generated_at`

当索引尚未建立时，返回 `ready=false`，而不是直接报错。

## 索引存储

索引写入当前项目目录下：

```text
.claude/code-index/
├── manifest.json
├── files.jsonl
├── symbols.jsonl
└── chunks.jsonl
```

其中：
- `manifest.json` 保存文件快照和 data file 摘要
- `files.jsonl` 保存文件级记录
- `symbols.jsonl` 保存 Go 符号记录
- `chunks.jsonl` 保存启发式 chunk 记录

## 技能说明

### `code-index-init`
用于初始化或重建当前项目索引。

### `code-index-refresh`
用于代码变更后刷新索引。

### `code-index-search`
用于在读源码前优先查询本地索引，缩小搜索范围。

## 本地验证

在插件目录内执行：

```bash
GOWORK=off go test ./...
GOWORK=off go build ./...
GOWORK=off go run ./cmd/code-index-mcp
```

## 在 Claude Code 中测试插件

可以通过插件目录直接测试：

```bash
zcode run --plugin-dir /Users/voidmind/GolandProjects/aitools/plugins/code-index-plugin
```

非交互最小验证示例：

```bash
zcode run --plugin-dir /Users/voidmind/GolandProjects/aitools/plugins/code-index-plugin \
  --allowedTools "mcp__plugin_code-index-plugin_code-index__get_code_index_status,mcp__plugin_code-index-plugin_code-index__build_code_index" \
  -p "先查看当前项目有没有索引，没有就为当前项目构建一个代码索引"
```

## 当前边界

第一版当前已支持：

- 文件级索引
- Go 符号索引
- 启发式 chunk 索引
- 索引构建与刷新
- 本地搜索与状态查询

第一版当前不做：

- 向量检索
- 远程索引服务
- 跨仓库联合索引
- 复杂调用图或语义分析
