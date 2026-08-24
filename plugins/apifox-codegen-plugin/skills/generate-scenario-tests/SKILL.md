---
name: generate-scenario-tests
description: 根据 Go controller/service 业务链路，通过 Apifox CLI 为已有接口创建或更新单接口测试用例。当用户要求生成场景化测试、补 Apifox 测试用例、按接口生成业务断言或继续补测试时使用。
---

# Generate Scenario Tests

围绕已识别的 endpoint 生成 Apifox `test-case`。本 skill 处理单接口业务用例；跨接口多步骤流程应使用 `test-scenario`，不要混用两种资源结构。

## 开始前

读取：

- `../../references/apifox-cli-rules.md`
- `../../references/scenario-test-rules.md`
- `../../references/apifox-testcase-assertion-example.md`

执行 `apifox --version`、`apifox whoami` 和 `apifox test-case --help`。CLI 未就绪时使用 `apifox-cli-setup`。

## 输入确认

1. 确认 `projectId`、分支和运行环境。
2. 确认 endpoint 来源：刚创建的接口、用户指定的接口，或当前代码扫描结果。
3. 对每个接口执行 `endpoint get`，确认 endpoint ID、真实 path、请求结构和响应结构。
4. 没有可用 endpoint 时停止，不编造测试。
5. 保持用户已确认的接口范围，不扩成全项目测试。

## 设计测试

逐个接口读取 controller/service：

1. 先写一个正向主场景，覆盖核心业务动作。
2. 从明确的业务分支中补关键异常和必要兜底。
3. 使用业务中文命名 `<接口或场景名>-<验证点>`。
4. 正向主场景不能只验证 HTTP 200，至少包含：
   - 成功语义断言，例如 `$.errNo == 0`
   - 一个行为或结构断言，例如分页值、状态变化、总数关系或关键字段存在
5. 常规校验优先使用可视化 `type: assertion`，复杂逻辑才使用 `customScript`。

## CLI 写入

1. 获取测试分类与现有用例：

```bash
apifox test-case category --project <projectId> --branch <branch>
apifox test-case list --project <projectId> --branch <branch> --endpoint <endpointId> --page-size 500
```

`categoryId` 是客户端可见性的关键字段，必须来自 `test-case category`，不能猜测。

2. 获取当前 schema：

```bash
apifox cli-schema get test-case-create
apifox cli-schema get test-case-update
```

3. 在临时目录中生成完整 JSON：
   - 包含 `name`、有效 `categoryId` 和 `apiDetailId`。
   - 显式写入 endpoint 的真实 `path`。
   - `requestBody.data` 为字符串，不是 JSON 对象。
   - `preProcessors`、`postProcessors` 使用 schema 当前定义的扁平结构。
   - processor 使用稳定 ID。
   - 可视化断言使用 `subject=httpCode|responseJson|responseText` 和 schema 支持的 comparison。
4. 校验后创建：

```bash
apifox cli-schema validate test-case-create --file <case.json>
apifox test-case create --project <projectId> --branch <branch> --file <case.json>
```

5. 创建后立即 `test-case get <caseId>`，确认 path、requestBody、断言和处理器实际保存。

## 更新已有用例

1. 先 `test-case get <caseId>` 获取完整结构。
2. 在完整结构上修改；不要把 update 当作 JSON Patch，不要丢失已有数组项。
3. 执行：

```bash
apifox cli-schema validate test-case-update --file <case-update.json>
apifox test-case update <caseId> --project <projectId> --branch <branch> --file <case-update.json>
apifox test-case get <caseId> --project <projectId> --branch <branch>
```

## 运行验证

创建或更新后至少运行一次。优先显式指定环境，并把报告写到临时目录：

```bash
apifox test-case run <caseId> \
  --project <projectId> \
  --branch <branch> \
  --environment <environmentId> \
  --reporters json \
  --out-dir <temporary-report-dir>
```

- `cli-schema validate` 和 `get` 成功不代表运行期正确。
- 运行失败时区分接口定义、环境地址、请求数据、业务响应和断言失败。
- 未经用户要求不要上传报告，也不要把临时报告提交到业务仓库。

## 输出

报告使用的项目/分支/环境、endpoint 与 case ID、生成的主场景和异常场景、分类、回读结果、运行结果、失败原因及刻意省略的无依据边界。
