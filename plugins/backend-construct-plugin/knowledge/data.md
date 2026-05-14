# Data

- `data` 只负责查询、写入、缓存访问等边界操作；若需求涉及 model、持久化实体或缓存对象结构，统一并入 `data` 处理。
- Service 层禁止直接拼 SQL；复杂查询仍应通过数据层封装。
- 计划中涉及 `data` 时，要检查 N+1、Join、缓存读写边界。
- 未命中 `data` 标签时，不默认追加 model/data 任务。

## 规则补充

- 若项目有固定目录约定，优先沿用 `data/mysql/`、`data/cache/` 或现有等价目录，不在 plan 阶段擅自改目录布局。
- Data 层默认只返回结果与 `error`，日志记录放在 service 或更上层；不要在 data 层吞错，也不要在 data 层主动打业务日志。
- 若项目存在 Redis key 常量集中维护位置，新增 key 应复用原有常量组织方式，不在 data 层内联硬编码。
- 若需求涉及 model 结构变化，应同步检查对应查询字段、唯一键、缓存序列化结构是否受影响。

## 最小示例

```go
package mysql

type FlowConfigData struct{}

var FlowConfigDataIns = &FlowConfigData{}

func (d *FlowConfigData) GetFlowConfigList(ctx *gin.Context) ([]*FlowConfig, error) {
    wb := mysql.AND(
        mysql.EQ("app_id", req.AppID),
        mysql.IsNotDeleted(),
    )
    return FlowConfigModelIns.GetBy(ctx, wb)
}
```
