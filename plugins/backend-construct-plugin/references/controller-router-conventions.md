# Controller Router Conventions

## Scope

本文件整理 controller / router 细规则，用于补充 `knowledge/controller-router.md` 与 `examples/controller-router-example.md`。

## Controller Layout

若项目已有约定，优先沿用：

- `controllers/http/<module>/controller.go`
- 一个模块一个 controller 文件，或沿用现有拆分方式

## Controller Signature

常见 handler 形态：

```go
func GetFlowConfig(ctx *gin.Context)
```

## Controller Flow

推荐保持线性结构：

1. 参数绑定
2. 绑定失败记录日志并返回
3. 调用 service
4. 统一返回成功或失败响应

若项目已有约定，优先复用：

- `ctx.ShouldBind(...)`
- `ctx.ShouldBindJSON(...)`
- `base.RenderJsonSucc(...)`
- `base.RenderJsonFail(...)`

## Router Layout

若项目已有约定，优先沿用：

- `router/http/<module>.go`
- 主入口 `router/http.go`
- 模块级 `Init(group *gin.RouterGroup)`

## Router Rules

- Router 只负责路由注册与中间件接线。
- 不在 router 层混入业务逻辑。
- 只在确实需要修改 `router` 时调整路径挂载或中间件绑定。

## Minimal Example

```go
func Init(group *gin.RouterGroup) {
    group.GET("/list", controller.GetFlowConfig)
    group.POST("/update", controller.UpdateFlowConfig)
}
```
