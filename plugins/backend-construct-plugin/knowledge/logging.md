# Logging

- 日志不单独形成标签；只有在命中相关层时，才在 plan 中显式检查日志要求。
- 若本次修改触及失败路径，应明确日志位置与关键字段。
- 计划中命中 `service`、`data`、`task` 时，要检查错误日志是否需要同步调整。

## 规则补充

- 若项目已有统一日志库，优先复用，不在计划中引入新的日志接口。
- 失败日志至少包含函数名、关键业务 ID、下游错误信息；成功日志只保留关键业务节点，不在 plan 中默认扩散为全量埋点。
- 若项目已有 Error / Warn / Info / Debug 的分级习惯，计划里应说明本次修改主要影响哪一级，而不是只写“补日志”。

## 最小示例

```go
zlog.Errorf(ctx, "[GetFlowConfig] query failed, app_id:%s, err:%v", req.AppId, err)
zlog.Infof(ctx, "[UpdateFlowConfig] update success, id:%d", id)
```
