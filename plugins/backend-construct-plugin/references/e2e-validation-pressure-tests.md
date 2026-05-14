# E2E Validation Pressure Tests

用于校验 `write-plans-with-construct` 是否默认先产出最高自然层级验证，而不是回退到单测优先。

## Baseline Failure 1

输入任务：新增一个查询接口，确认标签 `service + data + controller + router`。

错误产物特征：
- 验证步骤先列“为 service 写单测”“为 data 写单测”
- 没有先给接口请求到响应的 E2E/集成验证
- 把 E2E 放到最后作为“补充”

## Baseline Failure 2

输入任务：补一个定时任务入口，确认标签 `task + service`。

错误产物特征：
- 直接拆成多个函数级单测
- 没有先给任务入口触发到业务执行的集成验证
- 没有要求在诊断后回到任务链路验证

## Expected Pass Signal

正确产物必须满足：
- 先写命中层的最高自然层级验证
- 有自然业务链路时，默认先写 E2E 或集成验证
- 只有当大粒度验证失败、定位不清或天然无 E2E 路径时，才补小粒度测试
- 补了小粒度测试后，最终仍回到大粒度验证闭环

## Scenario 1: HTTP query flow

输入任务：新增查询接口，确认标签 `service + data + controller + router`

通过标准：
- 计划先给接口请求到响应的 E2E/集成验证
- 计划不会默认先列 `service` / `data` 单测
- 若补更小粒度测试，必须明确是因为接口级验证失败后需要进一步定位
- 最终仍要回到接口级验证闭环

## Scenario 2: Scheduled task flow

输入任务：新增定时任务入口，确认标签 `task + service`

通过标准：
- 计划先给任务入口触发到业务执行的集成验证
- 若补小粒度测试，必须显式说明是为定位失败原因
- 最终仍要回到任务链路验证

## Scenario 3: Natural no-E2E path

输入任务：只调整一个纯 `dto` 字段映射规则，确认标签 `dto`

通过标准：
- 计划可以从更小但仍真实的验证边界起步
- 计划不会虚构 HTTP 级 E2E
- 计划会说明当前命中层天然没有自然端到端路径，因此采用局部验证

## Red Flags

- “先把 `service` 单测补齐，再看是否需要集成验证”
- “通常顺便为每层都补一份单测”
- “E2E 最后再补”
- “先写几个小测试更稳妥”
- “虽然有真实链路，但从单测起手更方便”

以上均视为失败信号。
