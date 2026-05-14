---
name: generate-scenario-tests
description: This skill should be used when the user asks to "生成场景化测试", "补 Apifox 测试用例", "基于接口生成测试", "根据 controller/service 逻辑补测试场景", "只生成 Apifox 测试", or "这些接口已经有了，继续补测试".
version: 0.1.0
allowed-tools:
  - Read
  - Bash
  - Skill
  - AskUserQuestion
---

# Generate Scenario Tests

基于已识别或已写入的接口生成 Apifox 场景化测试用例。

## Use Scope

先确认当前上下文里已经具备可用的接口基础，再继续生成测试：

- 已从代码链路识别出接口。
- 已写入或已存在于 Apifox 的接口。
- 已明确接口范围与接口清单。

若当前没有可用接口基础，不要直接编造测试。先要求补齐接口范围，或先继续调用接口生成流程，再回到本 skill。

## Read Before Work

开始前先阅读：

- `references/scenario-test-rules.md`：场景组织方式、覆盖顺序、输出边界、命名约束、强断言要求。

把 reference 中的关键约束提升到执行步骤里，不要只在生成完成后才回想：

- 正向主场景不能只验证 `HTTP 200`
- 测试用例应显式写入接口真实 `apiPath`
- 后置断言优先使用 Apifox UI 可识别的 `type: assertion` 结构

## Real MCP Mapping

涉及 Apifox 测试用例写入时，优先使用以下真实工具：

- 项目结构：`getProjectSummary`
- 接口列表：`getStructureInfo`
- 接口详情：`getHttpEndpoint` / `readEntityDetails`
- 测试用例列表：`listTestCases`
- 测试用例详情：`getTestCase`
- 创建测试用例：`createTestCase`
- 更新测试用例：`updateTestCase`
- 删除测试用例：`deleteTestCase`

若需要环境信息辅助判断路径或模块地址，继续走通用 OpenAPI 路径：`listOpenApiEndpoints -> getOpenApiDetails -> executeOpenApi`。

## Input Checks

按以下顺序确认输入：

1. 确认接口来源。
   - 来自当前会话中已识别的接口。
   - 来自刚写入 Apifox 的接口。
   - 来自用户明确指定的现有 Apifox 接口。
   - 需要时用 `getStructureInfo`、`getHttpEndpoint` 或 `readEntityDetails` 二次确认接口 ID、`apiPath`、响应结构。

2. 确认生成范围。
   - 按当前已确认的接口范围继续。
   - 不擅自扩成全量接口测试。

3. 确认输出目标。
   - 默认将结果写入 Apifox 测试分类。
   - 第一版只输出 Apifox 测试用例。
   - 不默认输出 XMind、Markdown 或仓库内自动化测试文件。

4. 确认断言与路径策略。
   - 正向测试默认走强断言，不允许只产出“返回 200/请求成功”。
   - 正向主场景必须显式写入接口真实 `apiPath`，不允许让路径字段保持空白或依赖隐式推断。
   - 若写入后置断言，必须使用 Apifox UI 可识别的标准 `type: assertion` 结构；若当前工具能力无法写入该结构，明确报告阻断，不回退到非标准断言结构。

## Analysis Procedure

按接口逐个生成测试，不要只看字段定义。

1. 先回看接口入口与主链路。
   - 从 controller 确认参数入口、成功返回、失败分支入口。
   - 从 service 提炼业务主流程、关键状态判断、失败兜底。

2. 按业务链路整理测试场景。
   - 先整理用户真正关心的业务动作。
   - 再抽取主流程成立条件、关键异常、必要边界。
   - 不按“字段一个个校验”机械罗列场景。

3. 统一用业务中文命名场景。
   - 优先使用业务动作、业务结果、业务约束描述。
   - 不直接拿变量名、结构体名、方法名做场景名。

## Minimum Coverage

每个接口至少满足以下最低覆盖：

1. 至少生成一个正向主场景。
   - 覆盖主流程成功路径。
   - 体现接口的核心业务价值。
   - 默认不能只验证 `HTTP 200`。
   - 必须显式写入接口真实 `apiPath`。
   - 至少包含成功断言、行为断言、结构断言三层中的有效组合。

2. 结合代码链路补关键异常。
   - 参数合法但业务不满足。
   - service 中明确分支出来的失败路径。
   - controller 中显式返回的错误分支。

3. 仅补必要边界。
   - 只保留业务上有意义的边界。
   - 不把所有字段边界都机械展开。

## Scenario Writing Rules

生成单个接口的测试时，按以下顺序输出：

1. 主场景。
   - 先写成功主流程。
   - 明确前置条件、关键步骤、预期结果。
   - 显式使用接口真实 `apiPath`。
   - 写入后置断言时，必须使用标准 `type: assertion` 结构。

2. 关键异常。
   - 只覆盖代码中有明确业务语义的异常分支。
   - 优先覆盖 service 决策点与 controller 返回分支。

3. 失败兜底。
   - 对存在统一失败返回、幂等保护、状态不允许、资源不存在等场景，补必要兜底用例。
   - 不凭空扩写代码中不存在的异常模型。

## Naming Convention

采用业务中文命名，并优先使用以下格式：

- `<接口或场景名>-<验证点>`

示例方向：

- `创建订单-正常下单成功`
- `创建订单-库存不足返回失败`
- `查询账单-未找到记录时返回空结果`

## Output Boundary

严格控制输出边界：

- 只产出 Apifox 测试用例。
- 默认写入 Apifox 测试分类。
- 生成前若需要判断是否已存在同类用例，先用 `listTestCases`。
- 新建用例时用 `createTestCase`，补强用例时用 `updateTestCase`，读取现有结构时用 `getTestCase`。
- 正向主场景必须带真实 `apiPath` 与标准 `type: assertion` 断言结构。
- 不默认生成 XMind。
- 不默认生成 Markdown 中间文档。
- 不默认生成仓库内自动化测试代码。

若用户后续明确要求其他产物，再单独确认并处理；不要在首版默认附带生成。

## Response Style

输出保持短促，并明确列出：

- 基于哪些接口生成测试。
- 每个接口生成了哪些主场景与关键异常场景。
- 哪些场景来自 controller/service 明确分支。
- 是否仅输出到 Apifox 测试分类。
- 哪些边界被刻意省略，以及省略原因。
