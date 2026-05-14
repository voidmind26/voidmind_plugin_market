# DTO Example

```go
package flowconfig

type UpdateFlowConfigReq struct {
    Id      int64  `json:"id" form:"id"`           // 配置ID
    AppId   string `json:"app_id" form:"app_id"`   // 应用ID
    Version string `json:"version" form:"version"` // 版本号
}

type UpdateFlowConfigResp struct {
    Success bool `json:"success" form:"success"`   // 是否成功
}
```
