# Gateway Platform Web Console 操作指南

本文面向 `gateway-platform-plugin` 的日常使用者，重点说明如何通过 Web Console 配置本地路由、凭据 Key 和请求注入规则。

## 运行架构

当前 Go 重写版本由两部分组成：

- `server/`：Gin HTTP 网关、SQLite 持久化、配置 API 和内嵌前端资源。
- `frontend/`：Vue 3 Web Console，构建结果会复制到 `server/router/frontend_dist/`。

构建流程会分别生成 `bin/gateway-platform-mcp` 和 `bin/gateway-platform-http`。MCP 进程负责检查并启动独立 HTTP 程序，Web Console 和 `/gateway/<route>` 转发入口统一由 `127.0.0.1:18787` 提供。

## 入口地址

启动插件后，在浏览器打开：

```text
http://127.0.0.1:18787/app
```

如果页面打不开，可以先访问健康检查确认本地平台是否运行：

```text
http://127.0.0.1:18787/api/health
```

## 本地数据目录

Gateway Platform 的持久化数据和 HTTP 服务日志统一保存在以下可见目录：

```text
~/CodexData/gateway-platform-plugin/
├── gateway-platform.db
└── gateway-platform-http.log
```

- 该目录不位于插件安装缓存中，插件升级、重装或 cachebuster 更新不会再删除数据库。
- 目录权限固定为 `0700`，数据库和日志文件使用 `0600`，避免 Key 凭据被其他本机用户读取。
- 首次启动新版本时，如果当前插件目录存在旧版 `gateway-platform.db`，会自动迁移；目标目录已有数据库时不会覆盖。
- Dashboard 会显示实际数据目录以及 `READ / WRITE` 状态；`/api/health` 也会返回数据目录、数据库可写状态、插件版本、HTTP 进程 PID 和可执行文件路径，供 MCP 在插件升级后识别并接管过期实例。旧任务检测到更新版本后会直接复用，避免把 HTTP 服务降级。

如需显式更改位置，可在启动 Codex 前设置绝对路径：

```bash
export GATEWAY_PLATFORM_DATA_DIR=/your/visible/path/gateway-platform-plugin
```

Web Console 主要包含三个配置页面：

| 页面 | 用途 |
| --- | --- |
| `Routes` | 配置本地网关路径与真实上游地址。 |
| `Keys` | 配置 token、cookie、API key 等本地凭据。 |
| `References` | 检查注入规则是否引用了不存在的 Key，以及哪些 Key 没有被使用。 |

## 核心概念

一次完整配置通常由三部分组成：

1. `Route`：定义一个本地访问入口，以及它要转发到哪个上游服务。
2. `Key`：保存一个本地凭据值，例如 token、cookie、API key。
3. `Rewrite`：定义转发请求时，把哪个 Key 注入到 Header、Query 或 Cookie 的哪个字段里。

请求链路如下：

```text
本地请求 -> /gateway/<route> -> 匹配 Route -> 应用 Rewrite 注入 -> 转发到 Upstream URL
```

## 推荐配置流程

建议按下面顺序操作：

1. 先在 `Keys` 页面创建需要注入的凭据。
2. 再到 `Routes` 页面创建本地路由。
3. 保存 Route 后，在 Route 弹窗中配置 `Rewrites` 注入规则。
4. 到 `References` 页面检查是否有缺失引用。
5. 使用 `/gateway/<route>` 地址发起一次真实请求验证。

## 配置 Key

进入 `Keys` 页面，点击新建 Key。

### 字段说明

| 字段 | 填写方式 |
| --- | --- |
| `Name 名称` | 给凭据起一个容易识别的名字，例如 `ship-cookie`、`apifox-token`、`openapi-key`。 |
| `Value 凭据值` | 填写真实凭据值，例如 token、cookie 内容或 API key。 |
| `Description 说明` | 写清楚这个凭据服务于哪个系统或场景。 |
| `Source 来源` | 标记凭据来源，例如 `manual`、`local`、`imported`。 |

### 使用建议

- 一个 Key 只保存一个明确用途的凭据，不要把多个 token 混在一起。
- `Name` 推荐带业务前缀，方便在 Rewrite 中选择。
- 凭据变更后，直接编辑对应 Key 即可，引用它的 Rewrite 会使用最新值。

## 配置 Route

进入 `Routes` 页面，点击新建 Route。

### 字段说明

| 字段 | 填写方式 |
| --- | --- |
| `Name 名称` | 路由名称，例如 `ship`、`apifox`、`docs`。 |
| `Local Path 本地路径` | 本地匹配路径，例如 `/ship`。实际访问时需要加上 `/gateway` 前缀。 |
| `Upstream URL 上游地址` | 真实目标服务地址，例如 `https://example.com/api`。 |
| `Timeout (ms) 超时` | 请求超时时间，默认 `30000`。 |
| `Enabled 启用状态` | 开启后允许转发；关闭后命中该 Route 会返回禁用错误。 |
| `Description 说明` | 写清楚这条路由转发到哪个系统、给什么场景使用。 |

### Local Path 与访问地址

`Local Path` 只填写路由自身路径，例如：

```text
/ship
```

真实访问时使用：

```text
http://127.0.0.1:18787/gateway/ship
```

如果继续访问子路径：

```text
http://127.0.0.1:18787/gateway/ship/v1/envs
```

并且 Route 的 `Upstream URL` 是：

```text
https://example.com/api
```

则最终会转发到：

```text
https://example.com/api/v1/envs
```

### 匹配规则

Route 使用最长前缀匹配。

例如同时存在：

| Local Path | Upstream URL |
| --- | --- |
| `/ship` | `https://example.com/ship` |
| `/ship-admin` | `https://example.com/admin` |

访问 `/gateway/ship-admin/users` 时，会优先匹配 `/ship-admin`。

## 配置 Rewrite 注入规则

保存 Route 后，在 Route 编辑弹窗下方可以看到 `Rewrites 注入规则`。点击新增注入规则后，按顺序配置：

1. 选择 `Type 注入位置`。
2. 选择 `Preset 常见模式`。
3. 选择要引用的 `Key 凭据`。
4. 必要时修改 `Target name 目标字段` 和 `Template 模板`。
5. 点击保存规则。

### 注入位置

| Type | 作用 | 示例 |
| --- | --- | --- |
| `Header 请求头` | 把 Key 注入到请求头。 | `Authorization: Bearer xxx` |
| `Query 查询参数` | 把 Key 注入到 URL 查询参数。 | `?token=xxx` |
| `Cookie 注入` | 把 Key 注入为 Cookie。 | `ZYBIPSCAS=xxx` |

### 常见模式

| Preset | Type | Target name | Template | 适用场景 |
| --- | --- | --- | --- | --- |
| `Bearer Token` | Header | `Authorization` | `Bearer {{value}}` | 标准 Bearer 认证。 |
| `Raw Header Token` | Header | `Authorization` | `{{value}}` | 上游要求 Header 值直接等于 token。 |
| `Session Cookie` | Cookie | `ZYBIPSCAS` | `{{value}}` | 需要注入登录态 Cookie。 |
| `Custom Cookie` | Cookie | `SESSION` | `{{value}}` | 自定义 Cookie 名称。 |
| `Token Query` | Query | `token` | `{{value}}` | 上游通过 `token` 参数鉴权。 |
| `Custom Query` | Query | `query_token` | `{{value}}` | 自定义查询参数名。 |

### Template 规则

`Template` 用来决定最终注入值的格式，其中：

```text
{{value}}
```

会被替换成所选 Key 的真实值。

常见写法：

| Template | 最终效果 |
| --- | --- |
| `{{value}}` | 原样注入 Key 值。 |
| `Bearer {{value}}` | 在 Key 值前加 `Bearer `。 |
| `token={{value}}` | 拼成自定义格式字符串。 |

### Target name 规则

`Target name` 的含义取决于注入位置：

| Type | Target name 表示 |
| --- | --- |
| `Header` | 请求头名称，例如 `Authorization`、`X-API-Key`。 |
| `Query` | 查询参数名称，例如 `token`、`access_key`。 |
| `Cookie` | Cookie 名称，例如 `ZYBIPSCAS`、`SESSION`。 |

### Ordering 排序

高级设置中的 `Ordering` 用于控制多条 Rewrite 的顺序。大多数场景不需要修改，只有同一个 Route 下存在多条注入规则且顺序有要求时再调整。

## 常见配置示例

### 示例一：为上游接口注入 Bearer Token

目标：访问本地 `/gateway/apifox/...` 时，自动带上：

```text
Authorization: Bearer <token>
```

配置方式：

1. 在 `Keys` 新建 Key：
   - `Name`: `apifox-token`
   - `Value`: 填写真实 token
2. 在 `Routes` 新建 Route：
   - `Name`: `apifox`
   - `Local Path`: `/apifox`
   - `Upstream URL`: 上游 API 地址
   - `Enabled`: 开启
3. 在该 Route 的 `Rewrites` 新增规则：
   - `Type`: `Header 请求头`
   - `Preset`: `Bearer Token`
   - `Key`: `apifox-token`
   - `Target name`: `Authorization`
   - `Template`: `Bearer {{value}}`
4. 使用 `/gateway/apifox/...` 发起请求验证。

### 示例二：为页面接口注入登录态 Cookie

目标：访问本地 `/gateway/ship/...` 时，自动带上：

```text
Cookie: ZYBIPSCAS=<cookie-value>
```

配置方式：

1. 在 `Keys` 新建 Key：
   - `Name`: `ship-cookie`
   - `Value`: 填写真实 Cookie 值
2. 在 `Routes` 新建 Route：
   - `Name`: `ship`
   - `Local Path`: `/ship`
   - `Upstream URL`: Ship 上游地址
3. 在该 Route 的 `Rewrites` 新增规则：
   - `Type`: `Cookie 注入`
   - `Preset`: `Session Cookie`
   - `Key`: `ship-cookie`
   - `Target name`: `ZYBIPSCAS`
   - `Template`: `{{value}}`
4. 使用 `/gateway/ship/...` 发起请求验证。

### 示例三：通过 Query 参数注入 token

目标：访问本地 `/gateway/openapi/users` 时，转发到上游时自动变成类似：

```text
https://example.com/users?token=<token>
```

配置方式：

1. 在 `Keys` 新建 Key：
   - `Name`: `openapi-token`
   - `Value`: 填写真实 token
2. 在 `Routes` 新建 Route：
   - `Name`: `openapi`
   - `Local Path`: `/openapi`
   - `Upstream URL`: `https://example.com`
3. 在该 Route 的 `Rewrites` 新增规则：
   - `Type`: `Query 查询参数`
   - `Preset`: `Token Query`
   - `Key`: `openapi-token`
   - `Target name`: `token`
   - `Template`: `{{value}}`
4. 使用 `/gateway/openapi/users` 发起请求验证。

## 检查 References

完成 Route、Key、Rewrite 配置后，进入 `References` 页面检查引用状态。

重点关注两类信息：

| 类型 | 含义 | 处理方式 |
| --- | --- | --- |
| 缺失引用 | Rewrite 引用了不存在的 Key。 | 回到 Route 的 Rewrite 中重新选择 Key，或补建对应 Key。 |
| 未使用 Key | Key 已创建，但没有被任何 Rewrite 使用。 | 如果不再需要，可以删除；如果需要，补充 Rewrite 引用。 |

建议每次删除 Key 或调整 Rewrite 后，都到 `References` 页面确认一次。

## 验证配置是否生效

### 1. 验证 Route 是否命中

访问：

```text
http://127.0.0.1:18787/gateway/<local-path>
```

如果返回 `route not found`，说明路径没有匹配到 Route。检查：

- 是否带了 `/gateway/` 前缀。
- Route 的 `Local Path` 是否填写正确。
- Local Path 是否以 `/` 开头。

### 2. 验证 Route 是否启用

如果返回 `route disabled`，说明 Route 的 `Enabled` 是关闭状态。进入 `Routes` 页面打开后重试。

### 3. 验证注入是否正确

优先使用能回显请求信息的上游接口进行验证，或者在上游服务日志中确认：

- Header 是否包含目标字段。
- Query 是否包含目标参数。
- Cookie 是否包含目标名称。

如果注入值不符合预期，检查：

- Rewrite 是否选择了正确 Key。
- `Template` 是否包含 `{{value}}`。
- `Target name` 是否是上游实际要求的字段名。
- `References` 页面是否提示缺失引用。

## 日常维护建议

- 新增凭据时先建 Key，再建 Rewrite。
- 修改凭据值时只改 Key，不需要改 Route。
- 删除 Key 前先看 `References`，避免造成缺失引用。
- 一个 Route 可以配置多条 Rewrite，用于同时注入 Header、Query 和 Cookie。
- Route 暂时不用时优先关闭 `Enabled`，确认不再需要后再删除。
- 不要把真实凭据截图、数据库文件或日志提交到仓库。
