# Service

- `service` 负责业务编排、错误处理和日志记录。
- 方法签名需与 DTO 和调用链保持一致。
- 计划中命中 `service` 时，应列出逻辑修改点、依赖调用点和验证方式。
- 未命中 `service` 标签时，不在 plan 中默认调整业务逻辑。

## 规则补充

- 若项目已有模式，优先沿用 `service/<module>/service.go`、结构体方法模式或单例模式，不在 plan 阶段无关重构。
- Service 层禁止直接拼 SQL，禁止把复杂 Redis 命令细节直接写进业务逻辑。
- Service 层要显式检查是否会引入 N+1 查询或不必要的联表；若项目已有“禁止联表”约束，应在计划中明确保留。
- 失败路径日志应包含函数名、关键业务 ID、下游错误信息；不要只返回错误而完全不留上下文。

## 最小示例

```go
func (s *Service) GetFlowConfig(ctx *gin.Context, req *flowconfig.GetFlowConfigReq) (*flowconfig.GetFlowConfigResp, error) {
    list, err := mysql.FlowConfigDataIns.GetFlowConfigList(ctx, req.AppId)
    if err != nil {
        zlog.Errorf(ctx, "[GetFlowConfig] query failed, app_id:%s, err:%v", req.AppId, err)
        return nil, err
    }

    resp := &flowconfig.GetFlowConfigResp{List: make([]flowconfig.FlowConfigData, 0, len(list))}
    return resp, nil
}
```
