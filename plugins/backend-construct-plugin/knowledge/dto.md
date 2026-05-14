# DTO

- 若项目采用按模块组织 DTO，则文件按模块落在 `dto/<module>/`。
- Request / Response 结构体命名应直接对应接口意图。
- 只在需要修改输入输出结构时给 `dto` 打标签。
- DTO 变更计划必须说明是否影响 controller 参数绑定与 service 返回。

## 规则补充

- 字段名应与 tag 语义一致，再转成 PascalCase，例如 `user_id -> UserId`。
- 常见 ID 字段优先使用 `int64`；时间戳优先使用 `int64`；状态位按项目习惯使用 `int` 或 `string`。
- 若项目当前同时使用 `json` / `form` tag，则 DTO 新增字段默认保持双 tag 一致。
- 导出字段如项目已有行尾注释习惯，新增字段应保持一致，不要一半有注释一半没有。

## 最小示例

```go
package flowconfig

type UpdateFlowConfigReq struct {
    Id      int64  `json:"id" form:"id"`             // 配置ID
    AppId   string `json:"app_id" form:"app_id"`     // 应用ID
    Version string `json:"version" form:"version"`   // 版本号
}

type UpdateFlowConfigResp struct {
    Success bool `json:"success" form:"success"`     // 是否成功
}
```
