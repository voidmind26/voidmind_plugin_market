# DTO Conventions

## Scope

本文件整理 DTO 的细规则，用于补充 `knowledge/dto.md` 与 `examples/dto-example.md`。

## File Layout

若项目采用模块化 DTO 结构，优先沿用：

- `dto/<module>/`
- 包名与目录名一致
- 文件名与接口意图相关，而不是泛用 `common.go`

若目标项目已有不同布局，以现有布局为准。

## Naming Rules

- Request / Response 结构体名直接对应接口意图
- 字段名与 tag 语义一致，再转成 PascalCase
  - `user_id` -> `UserId`
  - `page_num` -> `PageNum`

## Tag Rules

若项目当前同时使用 `json` / `form` tag，则新增字段默认保持双 tag：

```go
UserId int64 `json:"user_id" form:"user_id"`
```

若项目只使用单一 tag，按现有项目规则走，不强推双 tag。

## Type Preferences

常见偏好：

- ID 字段：`int64`
- 时间戳：`int64`
- 状态位：`int` 或 `string`
- 复杂对象：显式嵌套 struct，不把 JSON 字符串当 DTO 主体

## Comment Style

若项目已有导出字段行尾注释习惯，新增 DTO 字段保持一致：

```go
AppId string `json:"app_id" form:"app_id"` // 应用ID
```

## Change Impact

修改 DTO 时同时检查：

- controller 参数绑定是否受影响
- service 返回结构是否受影响
- 接口响应兼容性是否需要额外说明
- 校验规则、可选字段和零值语义是否保持兼容
