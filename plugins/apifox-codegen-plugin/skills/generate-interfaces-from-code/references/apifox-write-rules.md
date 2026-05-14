# Apifox 写入规则

## 1. Path 三层拆分

在代码扫描阶段，同时维护三层路径概念，避免把扫描边界直接写进 Apifox：

1. `routePrefix`
   - 含义：用户指定的扫描边界，例如 `/user`、`/order`。
   - 用途：只用于筛选要不要进入分析。
   - 禁止：不要把它当成最终写入 path 的强制前缀。

2. `fullRoute`
   - 含义：从 router 实际还原出的完整路由，例如 `/api/user/profile`。
   - 用途：作为 router、controller、handler 的代码定位证据。
   - 禁止：不要直接把 fullRoute 原样写入 Apifox。

3. `apiPath`
   - 含义：写入 Apifox 的最终接口路径，只保留接口自身路径语义，例如 `/profile`。
   - 用途：Apifox `path` 字段唯一合法来源。
   - 要求：生成到 Apifox 时，`path = apiPath`，不要额外拼回 `routePrefix` 或 `fullRoute`。

简化理解：

- `routePrefix` 用来决定“扫哪里”
- `fullRoute` 用来证明“代码里在哪里”
- `apiPath` 用来决定“Apifox 里写什么”

## 1.1 模块与测试环境约束

在新增 Apifox 接口前，除了 path 三层拆分，还必须额外确认两件事：

1. 接口必须创建在目标项目的目标模块下，不能默认落到默认模块。
2. 测试环境中必须已经为该模块配置对应的模块地址；若没有模块地址，先补环境，再创建接口。

检查顺序：

1. 先通过项目结构确认 `moduleId` 是否存在。
2. 再读取项目环境列表，找到测试环境。
3. 检查测试环境 `baseUrls` 中是否存在该 `moduleId` 的地址。
4. 只有模块存在且测试环境模块地址存在时，才允许创建接口。

## 2. 写入顺序

严格使用“最小创建 -> 读取详情 -> 分步补全”的顺序，避免一次性提交过大 payload。

## 2.1 真实工具映射

优先使用以下 Apifox MCP 工具：

- 项目列表：`listAccessibleProjects`
- 项目结构：`getProjectSummary`
- 接口列表：`getStructureInfo`
- 接口详情：`getHttpEndpoint` / `readEntityDetails`
- 新建接口：`createHttpEndpoint`
- 更新接口：`updateHttpEndpoint`
- 删除接口：`deleteHttpEndpoint`

环境相关能力当前走通用 OpenAPI 路径：

1. `listOpenApiEndpoints` 查找环境端点
2. `getOpenApiDetails` 读取端点参数结构
3. `executeOpenApi` 执行环境查询或修改

### 第一步：最小创建

先创建接口骨架，只放最小必需字段：

- method
- path（只传 `apiPath`）
- title 或最小接口名

目标不是一次成型，而是先拿到可读、可更新的接口实体。

### 第二步：读取详情

创建成功后立即读取详情，确认：

- Apifox 返回的真实字段结构
- 当前接口 ID / 分类归属
- 后续补写 description、schema、example 时应落在哪些字段

不要跳过这一步。不要凭想象直接拼下一次大更新。

### 第三步：分步补全

按固定顺序补内容：

1. `description`
2. request schema
3. response schema
4. example

每一步都只做一类变更。任何一步失败，都回到该步继续缩圈，不把更多字段一起混入。

## 3. 422 缩圈规则

422 代表提交结构或字段值不被接口接受。排查时不要全量回滚后重写；要逐层缩圈。

### 缩圈起点

先确认最小创建可成功。若最小创建都失败，优先检查：

- method 是否合法
- `path` 是否误传了 `fullRoute` 或额外前缀
- 标题或分类等最小必填项是否缺失

### 缩圈顺序

最小创建成功后，每次只加一个维度：

1. description
2. request schema
3. response schema
4. example

一旦某一步 422，说明问题已缩到当前维度，不要继续叠加下一步。

### 当前维度内继续缩小

若 request 或 response schema 出错，再按层级继续拆：

- 先只提交顶层对象
- 再补一级字段
- 再补嵌套对象
- 最后补数组元素结构与示例值

若 example 出错，再拆成：

- 先空 example 或最小 example
- 再补一层字段
- 再补完整示例

### 缩圈目标

最终要定位到最小失败单元，例如：

- 某个字段类型名不合法
- 某个嵌套数组结构不被接受
- 某段 schema 没有先序列化为 JSON 字符串
- 某个 example 字段与 schema 不匹配

定位后优先修正；无法合理修正时，保留最小可用写入结果，不伪造复杂字段。

## 4. 复杂 schema 序列化规则

复杂 response schema 不要边拼边发。先在本地组织稳定结构，再一次序列化提交。

### 适用场景

以下场景都按“先本地组织，再序列化”处理：

- 多层嵌套对象
- 对象数组
- 列表项里再嵌套对象
- 通用响应包装内包 data / list / pageInfo
- 同时包含 nullable、枚举、示例等附加信息

### 推荐步骤

1. 在本地先整理清楚结构树。
2. 确认每个字段的：
   - 名称
   - 类型
   - 是否数组
   - 是否对象
   - 子字段
3. 结构稳定后，再序列化成 JSON 字符串。
4. 将该 JSON 字符串提交给 Apifox 对应字段。

### 禁止做法

- 一边猜字段，一边直接提交大对象
- 将未闭合、未成型的嵌套结构直接写入
- 将本地对象和字符串化对象混用

## 5. 结果要求

最终写入结果至少满足：

- `path` 只等于 `apiPath`
- 新增接口已落在目标模块，不在默认模块
- 测试环境已存在该模块地址
- 已按固定顺序完成写入或缩圈
- 复杂 schema 已先本地组织再序列化
- 422 时已能明确指出失败维度或失败字段范围
