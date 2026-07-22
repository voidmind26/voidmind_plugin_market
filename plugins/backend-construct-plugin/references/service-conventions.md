# Service Conventions

## Scope

本文件整理 service 相关的细规则，用于补充 `knowledge/service.md` 与 `examples/service-example.md`。

## File Layout

若项目已有约定，优先沿用：

- `service/<module>/service.go`
- 公共辅助逻辑放在该模块内或既有 `funcs.go`，不要为了当前需求额外发明层次

## Common Shapes

常见形态有两种：

1. 结构体方法模式
2. 单例实例模式

若项目已有 `Instance = &Service{}` 风格，新增逻辑优先复用，不在当前业务修改中切换模式。

## Signature Pattern

常见方法签名：

```go
func (s *Service) MethodName(ctx *gin.Context, req *dto.Req) (*dto.Resp, error)
```

若项目已有不同上下文类型或请求对象类型，以现有项目为准。

## Hard Rules

- Service 负责业务编排，不负责直接拼 SQL。
- Service 不直接承载复杂 Redis 命令细节，应通过 data/cache 层封装。
- 必须检查是否引入 N+1 查询。
- 若项目已有“禁止联表”约束，实现中必须保留，不因当前需求放宽。

## Logging

失败路径建议保留：

- 函数名
- 关键业务 ID
- 下游错误信息

示例：

```go
zlog.Errorf(ctx, "[GetFlowConfig] query failed, app_id:%s, err:%v", req.AppId, err)
```

## Response Assembly

若 service 负责拼装响应对象：

- 保持 DTO 字段与上游返回结构一致
- 不在响应组装阶段顺手引入额外业务逻辑
- 复杂映射应单独说明其来源与验证方式
