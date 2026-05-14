# Controller Router

- `controller` 只做参数绑定、service 调用与统一响应。
- `router` 只做路由注册与中间件绑定。
- `controller+router` 常一起出现，但可只修改其一。
- 未命中 `controller` 或 `router` 标签时，不追加 HTTP 接入层任务。

## 规则补充

- 若项目已有目录约定，优先沿用 `controllers/http/<module>/controller.go` 与 `router/http/<module>.go` 这类结构，不在 plan 阶段擅自重排接入层目录。
- Controller 层优先保持“绑定参数 -> 调 service -> 统一返回”的线性结构；不要把业务判断堆进 controller。
- 若项目已有 `ShouldBind / ShouldBindJSON`、`RenderJsonSucc / RenderJsonFail`、统一错误响应方式，应在计划中显式复用。
- Router 层优先保持 `Init(group *gin.RouterGroup)` 或项目既有等价模式；只在命中 `router` 标签时安排路由注册任务。

## 最小示例

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
```
