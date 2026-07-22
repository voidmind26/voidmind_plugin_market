# Task Registration Example

## Job File

```go
package command

const JobRefreshConfigCycle = 5

func JobRefreshConfig(ctx *gin.Context) error {
    if !redisLock(ctx, "job:refresh_config:lock", JobRefreshConfigCycle*60) {
        return nil
    }

    if err := service.Instance.RefreshConfig(ctx); err != nil {
        zlog.Errorf(ctx, "[JobRefreshConfig] failed, err:%v", err)
        return nil
    }

    zlog.Infof(ctx, "[JobRefreshConfig] success")
    return nil
}
```

## Registration

```go
func startCycle(engine *gin.Engine) {
    cycleJob := command.InitCycle(engine)
    cycleJob.AddFunc(time.Minute*JobRefreshConfigCycle, command.JobRefreshConfig)
    cycleJob.Start()
}
```

## Notes

- 任务失败时返回 `nil`，避免同类错误被周期调度重复放大。
- 多实例部署时优先补分布式锁或等价互斥方案。
- 若任务只改入口 wiring，不默认追加 controller/router 代码。
