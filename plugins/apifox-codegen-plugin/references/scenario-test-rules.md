# 场景化测试生成规则

## 1. 资源边界

- 单接口业务验证写入 `test-case`。
- 登录 -> 创建 -> 查询等跨接口流程写入 `test-scenario`。
- 本插件默认生成单接口 `test-case`，不要把 test-scenario 的步骤结构写进 test-case。
- 第一版不默认生成 XMind、Markdown 或仓库内自动化测试代码。

## 2. 场景组织

- 按业务链路组织，不按字段清单平铺。
- 先从 controller 确认参数与返回，再从 service 提取主流程、状态分支和失败兜底。
- 每个接口至少有一个正向主场景。
- 只补代码中有明确依据的异常和边界。
- 使用业务中文命名 `<接口或场景名>-<验证点>`。

## 3. 最低断言

正向主场景不能只验证 HTTP 200，至少覆盖：

1. 成功语义，例如 `$.errNo == 0`。
2. 一个行为或结构结果，例如分页值、排序、状态变化、总数关系、数组长度或关键字段存在。

常规断言优先使用 `type: assertion`：

- 状态码：`subject=httpCode`
- JSON 字段：`subject=responseJson`
- 文本包含：`subject=responseText`
- 相等比较：`comparison=equal`

复杂数组筛选、条件计算或二次请求才使用 `customScript`。

## 4. CLI 创建流程

```bash
apifox endpoint get <endpointId> --project <projectId> --branch <branch>
apifox test-case category --project <projectId> --branch <branch>
apifox test-case list --project <projectId> --branch <branch> \
  --endpoint <endpointId> --page-size 500
apifox cli-schema get test-case-create
apifox cli-schema validate test-case-create --file <case.json>
apifox test-case create --project <projectId> --branch <branch> --file <case.json>
apifox test-case get <caseId> --project <projectId> --branch <branch>
```

- `categoryId` 必须来自 `test-case category`，不能使用猜测值。
- `apiDetailId` 绑定真实 endpoint。
- `path` 显式使用 endpoint 的真实 `apifoxPath`。
- `requestBody.data` 必须是字符串。
- processor 使用 `{id,type,data,defaultEnable,enable}` 的当前 schema 结构。

## 5. 更新流程

更新前必须 `get` 完整测试用例；基于完整结构修改并用 `test-case-update` schema 校验。update 不是 JSON Patch，不会自动合并 `preProcessors` 或 `postProcessors`。

## 6. 运行验证

创建或更新后至少执行一次 `test-case run`，建议显式指定环境并生成临时 JSON 报告。

验证失败时分别判断：

- endpoint 定义是否正确
- 环境模块地址是否正确
- 请求参数/Body 是否有效
- 业务响应是否符合预期
- assertion 或 customScript 是否写错

不要因为 schema 校验和回读成功就宣称测试可运行。
