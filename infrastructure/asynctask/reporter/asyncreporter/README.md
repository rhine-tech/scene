# Async Reporter

`asyncreporter` 是 `asynctask.TaskReporter` 的异步 decorator。
它不改变业务调用方式，只把同步 reporter 包一层有界缓冲和后台 worker。

## 为什么需要

任务执行是主路径，状态上报是可观测性能力。
如果底层 reporter 写 Redis、MongoDB、SQL 或远程服务时卡住，不应该直接拖住 worker 的业务执行。

`AsyncReporter` 的目标是：

- 让 `ReportTaskEvent` 快速返回。
- 用有界 buffer 防止内存无限增长。
- 通过 drop policy 明确 buffer 满时的降级行为。
- 通过 `Flush` / `Stop` 支持 worker 退出前尽量写完已接收事件。

## 使用方式

```go
memory := memoryreporter.NewMemoryReporter()
reporter := asyncreporter.New(memory, asyncreporter.Config{
	BufferSize:    4096,
	Workers:       1,
	ReportTimeout: time.Second,
	OnFull:        asyncreporter.DropNewest,
})

registry.Register[asynctask.TaskReporter](reporter)
```

业务 handler 仍然只依赖 `asynctask.TaskReporter`：

```go
type Handler struct {
	reporter asynctask.TaskReporter `aperture:""`
}
```

## Drop Policy

`DropNewest` 是默认策略。
当 buffer 满时丢弃新事件，保护 worker 不被 reporter 拖住。

`DropOldest` 会丢弃一个已缓冲的旧事件，再尝试写入新事件。
适合更关心最新状态的场景。

`Block` 会等待 buffer 可写，直到调用方 context 取消。
只有当任务状态上报必须可靠时才建议使用；否则可能让 reporter 反向拖慢 worker。

## 生命周期

`ReportTaskEvent` 会懒启动后台 worker。
如果希望服务退出时尽量写完事件，应在 void app 或其他 lifecycle 中调用：

```go
_ = reporter.Flush(ctx)
_ = reporter.Stop(ctx)
```

`Flush` 等待已接受事件被底层 reporter 处理完成。
`Stop` 会停止后台 worker，并在退出前 drain buffer。

注意：`Stop` 无法强杀一个已经卡在底层 reporter 的 goroutine。
因此底层 reporter 应该尊重 context，`ReportTimeout` 也应该设置为合理值。

## 错误语义

`AsyncReporter` 默认把底层 reporter 错误视为观测失败，不把错误返回给业务调用方。
如果需要记录错误，可以配置 `OnError`。

buffer 满导致丢弃事件时，可以通过 `OnDrop` 或 `Dropped()` 观察。

```go
reporter := asyncreporter.New(target, asyncreporter.Config{
	OnError: func(err error, event asynctask.TaskEvent) {
		// log or metrics
	},
	OnDrop: func(event asynctask.TaskEvent) {
		// metrics
	},
})
```

## 适用边界

Async reporter 适合保护 worker 不被 reporter 写入卡住。
它不是可靠事件队列，也不保证任务状态 100% 不丢。

如果业务要求任务状态绝对可靠，应该使用可靠的持久化 reporter，并让业务接受上报失败会影响任务流程这一事实。
