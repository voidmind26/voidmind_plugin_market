# Task Logging Conventions

## Scope

本文件整理 task 与 logging 的细规则，用于补充 `knowledge/task.md`、`knowledge/logging.md` 与 `examples/task-logging-example.md`。

## Task Layout

若项目已有约定，优先沿用：

- `controllers/command/job_<name>.go`
- 任务注册入口如 `router/command.go` 或现有等价位置

## Task Rules

- 周期任务通常配套周期常量，如 `JobRefreshConfigCycle = 5`。
- 多实例部署下，若任务存在重复执行风险，优先考虑分布式锁或等价互斥方案。
- 若项目已有约定，任务失败返回 `nil`，避免调度器重复放大同类错误。
- 若任务存在时间窗口限制，应在计划中明确说明，而不是靠隐式行为。

## Registration Pattern

常见注册形态：

```go
func startCycle(engine *gin.Engine) {
    cycleJob := command.InitCycle(engine)
    cycleJob.AddFunc(time.Minute*JobRefreshConfigCycle, command.JobRefreshConfig)
    cycleJob.Start()
}
```

## Logging Rules

若项目已有统一日志库，优先复用，如 `zlog`。

建议按语义选择级别：

- `Error`：影响主流程或导致失败
- `Warn`：降级、重试、兜底
- `Info`：关键业务节点
- `Debug`：排查问题时的细节

## Logging Format

若项目已有风格，优先复用函数名前缀：

```go
zlog.Errorf(ctx, "[JobRefreshConfig] failed, err:%v", err)
zlog.Infof(ctx, "[JobRefreshConfig] success")
```

失败日志优先带：

- 函数名
- 关键业务 ID / key
- 下游错误信息
