---
name: generate-interfaces-from-code
description: This skill should be used when the user asks to "从代码生成 Apifox 接口", "根据 Go 项目代码生成 Apifox 接口", "扫描 Go 路由生成接口文档", "按路由前缀补 Apifox 接口", "把现有路由整理到 Apifox", or "全量扫描接口生成 Apifox 文档".
version: 0.1.0
allowed-tools:
  - Read
  - Bash
  - Skill
  - AskUserQuestion
---

# Generate Interfaces From Code

从 Go Web 代码抽取接口信息，并按受控顺序写入 Apifox。

## Use Scope

先确认项目属于 Go Web 路由组织，再继续扫描。

- 默认只按用户给定的路由前缀扫描，例如 `/user`、`/order`。
- 只有用户明确确认“全部接口”时，才允许全量扫描。
- 若项目结构明显不是 router -> controller -> service -> dto 这类 Go Web 组织，直接说明当前 skill 不接管。

## Read Before Work

开始扫描前，先阅读以下参考文件，并把其中的阻断式约束带入主流程，不要只停留在“知道有这份 reference”：

- `references/route-scan-rules.md`：扫描边界、链路取证、保守降级规则。
- `references/apifox-write-rules.md`：Apifox 写入顺序、模块归属、测试环境模块地址校验、422 缩圈、复杂 schema 提交规则。

若未完成 reference 中要求的前置确认，不进入接口创建。

## Real MCP Mapping

涉及 Apifox 写入时，优先使用以下真实工具：

- 项目列表：`listAccessibleProjects`
- 项目结构：`getProjectSummary`
- 接口概要列表：`getStructureInfo`
- 读取接口详情：`getHttpEndpoint` / `readEntityDetails`
- 创建接口：`createHttpEndpoint`
- 更新接口：`updateHttpEndpoint`
- 删除接口：`deleteHttpEndpoint`
- 通用扩展能力发现：`listOpenApiEndpoints`
- 通用扩展能力详情：`getOpenApiDetails`
- 通用扩展能力执行：`executeOpenApi`

如果固定工具覆盖不了目标动作（例如环境管理），再走 `listOpenApiEndpoints -> getOpenApiDetails -> executeOpenApi`。

## Scan Procedure

按以下顺序执行：

1. 确认扫描范围。
   - 已给前缀：按该前缀扫描。
   - 未给前缀：先追问。
   - 用户要求“全部接口”：先说明是全量扫描，再取得明确确认。

2. 从 router 开始取证。
   - 识别命中的 method、fullRoute、handler。
   - 将 fullRoute 仅作为代码侧定位依据，不直接写入 Apifox。

3. 继续追 controller -> service -> dto。
   - controller：提取接口名、参数绑定、响应返回入口、注释说明。
   - service：提取业务语义、状态分支、核心字段含义。
   - dto：提取请求结构、响应结构、字段类型、json 标签、示例线索。

4. 为每个接口形成最小结构化结果。

```json
{
  "method": "GET",
  "fullRoute": "/api/user/profile",
  "apiPath": "/profile",
  "interfaceName": "获取用户资料",
  "request": {},
  "response": {},
  "descriptionDraft": "根据代码行为整理出的说明草案"
}
```

5. 缺少完整注释时做有限推断。
   - 允许根据命名、DTO 字段、返回包装、错误码约定推断接口名和说明草案。
   - 不能合理确定的信息不要伪造；保留为空、标记待确认，或在说明中显式写出不确定点。

## Extraction Rules

抽取结果至少覆盖以下字段：

- `method`
- `apiPath`
- `interfaceName`
- `request`
- `response`
- `descriptionDraft`

遵守以下约束：

- 扫描时可以保留 `fullRoute` 作为证据。
- 生成到 Apifox 时，`path` 只保留 `apiPath`。
- `apiPath` 不携带用于扫描的公共前缀。
- 接口说明优先取注释；缺注释时，再依据代码约定整理草案。

## Apifox Write Procedure

严格按以下顺序写入：

1. 先做创建前强检查。
   - 用 `listAccessibleProjects` 确认目标项目。
   - 用 `getProjectSummary` 确认接口应落到哪个目标模块，并拿到 `moduleId`。
   - 若需要确认接口是否已存在，可用 `getStructureInfo(entityType=endpoint)` 或 `readEntityDetails`。
   - 用环境相关通用能力检查测试环境：先 `listOpenApiEndpoints` 找环境端点，再 `getOpenApiDetails`，最后 `executeOpenApi` 读取环境列表与详情，确认测试环境中已为该 `moduleId` 配置专属地址。
   - 上述任一条件不满足时，停止创建，并明确回报阻断原因。

2. 最小创建。
   - 用 `createHttpEndpoint` 创建接口骨架。
   - 只提交最小可成立字段：`method`、`name`、`path`、`type=http`、`moduleId`。
   - `path` 仅传 `apiPath`。
   - 创建时显式使用已确认的目标模块，不允许默认落到默认模块。

3. 读取详情。
   - 创建成功后立即用 `getHttpEndpoint` 或 `readEntityDetails` 读取接口详情。
   - 以 Apifox 当前返回结构为准，确认后续补全所需字段位置。

4. 分步补全。
   - 用 `updateHttpEndpoint` 先补 `description`。
   - 再补请求 schema。
   - 再补响应 schema。
   - 最后补 example。

5. 复杂 response schema 先本地组织，再序列化提交。
   - 先在本地构造完整对象结构。
   - 确认字段层级、数组元素、嵌套对象都稳定后，再序列化成 JSON 字符串提交给 `updateHttpEndpoint`。

## 422 Narrowing

出现 422 时，不要一次性重写整份 payload。按“缩圈”方式定位：

1. 回到最小创建版本，确认接口骨架可成功。
2. 只增加一个维度后再次提交，顺序固定为：
   - `description`
   - request schema
   - response schema
   - example
3. 一旦某一步失败，停在该层继续缩小。
   - response 太复杂：先只提交顶层对象，再逐层补字段。
   - 数组对象失败：先提交空数组元素结构，再逐字段补全。
   - schema 疑似格式问题：先提交最小合法对象，再恢复附加字段。
4. 记录最终失败字段或层级；能修则修，不能合理确定则暂时留空，不伪造内容。

## Output Style

输出保持短促，明确列出：

- 扫描范围
- 命中的接口数
- 每个接口的 method + apiPath + 接口名
- 目标模块确认结果
- 测试环境模块地址校验结果
- 因前置校验失败而被阻断的接口（若有）
- 已写入项与待确认项
- 422 缩圈定位结果（若发生）
