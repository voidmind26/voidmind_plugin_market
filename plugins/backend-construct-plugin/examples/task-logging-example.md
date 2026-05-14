# Task Logging Example

```go
const JobRefreshConfigCycle = 5

func JobRefreshConfig(ctx *gin.Context) error {
    if err := service.Instance.RefreshConfig(ctx); err != nil {
        zlog.Errorf(ctx, "[JobRefreshConfig] failed, err:%v", err)
        return nil
    }
    zlog.Infof(ctx, "[JobRefreshConfig] success")
    return nil
}
```
