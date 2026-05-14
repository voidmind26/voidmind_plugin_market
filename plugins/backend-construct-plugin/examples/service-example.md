# Service Example

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
