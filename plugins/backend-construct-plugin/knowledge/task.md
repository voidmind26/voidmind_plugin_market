# Task

- `task` 标签仅用于定时任务、周期任务、任务中心或后台 worker 入口变更。
- 命中 `task` 时，应明确是否同时涉及 `service`。
- 普通接口请求链路中的异步处理，不因“有后台动作”就默认标成 `task`。
- 不因为存在定时任务背景就默认补 controller/router 计划。

## 规则补充

- 若项目已有任务目录约定，优先沿用 `controllers/command/job_<name>.go` 或现有等价结构，不在计划中凭空发明新入口布局。
- 多实例部署下若任务有并发重复执行风险，应显式考虑分布式锁或等价互斥方案。
- 若项目已有“任务失败返回 nil 避免重复报错”之类运行规则，应在计划中保留，不要遗漏运行期约束。
- 任务计划应单独说明：调度入口、执行周期、是否受时间窗口限制、是否复用现有 service。

## 最小示例

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
