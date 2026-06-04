# Memory Reporter / Inspector 设计说明

`memoryreporter` 是 `asynctask.TaskReporter` 和 `asynctask.TaskQueueInspector` 的内存实现。
它用于本地开发、测试、单进程 worker 或轻量级管理页面，不用于跨进程共享任务状态。

默认导出的是 `v2`：

```go
reporter := memoryreporter.NewMemoryReporter()
```

如果需要显式引用版本：

```go
v1reporter.NewMemoryReporter()
v2reporter.NewMemoryReporter()
```

## 目标

Memory reporter 只保存每个 task 的最新 `TaskEvent` 快照，而不是保存完整事件流。
它解决的是“查看当前任务状态”的问题，不解决审计日志、历史回放或分布式状态共享问题。

它同时实现两个接口：

- `TaskReporter`：worker 或 queue backend 上报任务状态、重试次数、进度、错误。
- `TaskQueueInspector`：管理端查询任务列表、单个任务、队列统计。

这两个接口拆开是为了保持使用方式稳定：

- worker 只依赖 reporter 上报。
- 管理接口只依赖 inspector 查询。
- 未来 Redis、MongoDB、SQL 或远程 reporter 可以替换 memory 实现，不影响业务 handler 的调用方式。

## 为什么保留 v1 和 v2

`v1` 是早期的 `sync.Map + atomic.Pointer` 实现。
它保留在代码中主要用于对比、回归验证和未来特殊场景参考，默认不使用。

`v2` 是当前默认实现，使用分片 map：

- 固定 64 个 shard。
- 每个 shard 内部是强类型 `map[string]TaskEvent`。
- 每个 shard 有独立 `RWMutex`。
- 上报只锁 taskID 所在的一个 shard，不锁全局。

保留两个版本的原因是二者代表了不同优化方向：

- `v1` 偏向 lock-free/CAS 思路，单个 task 的快照用 atomic pointer 更新。
- `v2` 偏向数据结构简单和写入热路径低分配，使用分片锁降低竞争。

当前 benchmark 显示，`v2` 对这个场景更合适。

## v1 设计

`v1` 的核心结构是：

- `sync.Map` 存 taskID 到 entry 的映射。
- entry 内部用 `atomic.Pointer[TaskEvent]` 保存最新快照。
- 更新时通过 CAS 合并事件。

优点：

- 没有全局互斥锁。
- 已存在 task 的更新不需要锁 map。
- 适合 key 高度分散、读多写少、生命周期复杂的场景。

缺点：

- 内部需要 `any` 类型断言。
- 每次成功上报都会生成新的 `TaskEvent` 快照，增加 GC 压力。
- CAS 循环和删除保护逻辑比普通 map 更难维护。
- 对当前 reporter 的“最新状态覆盖”场景并不比分片锁更快。

## v2 设计

`v2` 的核心结构是：

- `shards [64]memoryShard`
- `memoryShard.events map[string]TaskEvent`
- `memoryShard.mu sync.RWMutex`

上报流程：

1. trim 并校验 `TaskID`。
2. 根据 FNV-1a hash 找到 shard。
3. 锁定该 shard。
4. 读取当前事件并合并新事件。
5. 写回 map。
6. 按需触发惰性清理。

查询流程：

- `GetTask` 只读单个 shard。
- `GetQueueStats` 逐 shard 读锁统计。
- `ListTasks` 逐 shard 拷贝快照，再按 `UpdatedAt desc, TaskID asc` 排序分页。

这个查询不是严格的全局一致快照。
它是逐 shard 的弱一致快照，查询期间不会长时间阻塞所有上报。
对 memory reporter 来说这是合理取舍；如果需要强一致查询，应使用持久化 inspector。

## 字段合并语义

Memory reporter 接收的是局部事件，不要求每次上报都包含完整状态。

合并规则：

- `Queue`、`TaskType`、`Key`、`Status`、`Message` 非空才覆盖旧值。
- `Attempt` 为非 0 才覆盖旧值，避免进度事件把 retry attempt 重置为 0。
- `Progress` 为非 0 才覆盖旧值，初始值允许保持 0。
- `Error` 总是覆盖，允许用空字符串清除旧错误。
- `UpdatedAt` 总是使用新事件时间。

同时，`v2` 会忽略乱序事件：

```go
if event.UpdatedAt.Before(current.UpdatedAt) {
	return nil
}
```

这样可以避免异步上报、重试或网络抖动导致旧状态覆盖新状态。

## OOM 防护

Memory reporter 不能无限保留 task 状态，否则长时间运行的 worker 或管理服务会有 OOM 风险。

因此 v1/v2 都支持相同的配置：

```go
type Config struct {
	EventTTL        time.Duration
	MaxTasks        int
	CleanupInterval time.Duration
}
```

默认值：

- `EventTTL`: 24 小时。
- `MaxTasks`: 10000。
- `CleanupInterval`: 1 分钟。

负数表示关闭对应限制。

清理策略：

- 不启动后台 goroutine，避免额外生命周期管理。
- 上报时按 `CleanupInterval` 惰性触发清理。
- 超过 `MaxTasks` 时同步淘汰最旧任务。
- 淘汰时按 `UpdatedAt asc, TaskID asc` 排序。
- 删除前会再次检查快照时间，避免误删清理期间刚更新过的 task。

这个策略的目标是保护进程内存，而不是保证所有观测事件都能永久保存。
在极端 overflow 情况下，memory reporter 会优先牺牲旧观测数据来保证进程存活。

## Benchmark 结论

当前本机 benchmark：

```text
v1: sync.Map + atomic.Pointer
v2: sharded map + RWMutex
```

在并发上报场景里，v2 更快且无分配。

原因：

- v2 直接在 shard map 中覆盖 `TaskEvent`，没有每次上报的指针快照分配。
- v2 没有 `sync.Map` 的 `any` 类型断言和复杂 entry 状态。
- 同一个 task 的更新串行化，字段合并逻辑简单且可预测。
- 分片锁把竞争限制在单个 shard，而不是全局锁。

因此默认使用 v2。

## 使用建议

本地开发、测试、单进程 worker：

```go
asynctask.MemoryReporter{}
```

需要显式配置 retention：

```go
asynctask.MemoryReporter{
	Config: memoryreporter.Config{
		EventTTL:        6 * time.Hour,
		MaxTasks:        5000,
		CleanupInterval: time.Minute,
	},
}
```

分布式 worker、跨进程状态查询、需要长期保留任务记录：

- 不要使用 memory reporter。
- 使用 Redis、MongoDB、SQL 或远程服务实现 `TaskReporter` / `TaskQueueInspector`。
