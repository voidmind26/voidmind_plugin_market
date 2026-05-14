# Controller Router Example

```go
func GetFlowConfig(ctx *gin.Context) {
    var req flowconfig.GetFlowConfigReq
    if err := ctx.ShouldBind(&req); err != nil {
        zlog.Errorf(ctx, "[GetFlowConfig] bind failed, err:%v", err)
        base.RenderJsonFail(ctx, err)
        return
    }

    resp, err := service.Instance.GetFlowConfig(ctx, &req)
    if err != nil {
        base.RenderJsonFail(ctx, err)
        return
    }
    base.RenderJsonSucc(ctx, resp)
}

func Init(group *gin.RouterGroup) {
    group.GET("/list", controller.GetFlowConfig)
}
```
