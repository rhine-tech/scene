# Async Task Queue

This document describes the queue-style async task system under `scene/infrastructure/asynctask`.

## Overview

The async task module now contains two different async execution models:

- `TaskDispatcher`
  - A simple task pool abstraction.
  - Suitable when the caller only needs "run this function in background".
  - Good for lightweight fire-and-forget work inside a single process.

- `TaskQueuePublisher` + `TaskQueueConsumer`
  - A queue abstraction with queue-level concurrency control.
  - Suitable when tasks need to be isolated by logical queue, retried, or backed by a real message queue.
  - Good for tasks such as cache warmup, third-party API synchronization, or rate-limited background jobs.

Use `TaskDispatcher` when you only need local background execution.

Use the queue abstraction when you need any of the following:

- queue-level concurrency limits
- retry on failure
- pluggable queue backend
- better workload isolation between task types
- future support for cross-process consumption

## Directory Layout

The module is organized by execution model:

```text
infrastructure/asynctask/
  task.go
  queue.go
  errors.go
  queue.md
  factory/
  taskpool/
    ants/
    tuna/
    cron.go
    common.go
  queue/
    memoryqueue/
    rabbitmq/
    redisstream/
    common.go
```

- `taskpool/` contains task-pool-based implementations.
- `queue/` contains queue-based implementations.

## Core Interfaces

The queue API is defined in `queue.go`.

### `TaskQueuePublisher`

Responsible for publishing tasks into a queue backend.

```go
type TaskQueuePublisher interface {
	scene.Named
	Publish(ctx context.Context, task *QueueTask) (*QueueTask, error)
}
```

### `TaskQueueConsumer`

Responsible for registering queues, registering handlers, and consuming queue tasks.

```go
type TaskQueueConsumer interface {
	scene.Named
	RegisterQueue(queue string, config TaskQueueConfig) error
	RegisterHandler(queue string, taskType string, handler TaskQueueHandler) error
}
```

### `TaskQueueHandler`

Responsible for handling a single queue task.

```go
type TaskQueueHandler interface {
	HandleTask(ctx context.Context, task *QueueTask) error
}
```

For simple use cases, use `TaskQueueHandlerFunc`.

```go
type TaskQueueHandlerFunc func(ctx context.Context, task *QueueTask) error
```

## QueueTask

`QueueTask` is the unified task model passed between publisher and consumer.

Important fields:

- `ID`
  - Unique task identifier.
  - Auto-generated when empty.

- `Queue`
  - Logical queue name.
  - Used to isolate workloads and apply queue-level concurrency.

- `Type`
  - Task type inside the queue.
  - Used by the consumer to route the task to a handler.

- `Payload`
  - Task body as raw bytes.
  - Usually JSON.

- `Headers`
  - Optional metadata for routing or tracing.

- `Delay`
  - Requested delay before execution.

- `Timeout`
  - Timeout passed into the task handler context.

- `MaxRetry`
  - Task-level retry override.
  - If zero, the queue config is used.

- `Attempt`
  - Current retry count.

## Queue Config

Each logical queue is configured through `TaskQueueConfig`.

```go
type TaskQueueConfig struct {
	Concurrency int
	MaxRetry    int
	RetryDelay  time.Duration
	BufferSize  int
}
```

Meaning:

- `Concurrency`
  - Maximum number of workers consuming this logical queue.

- `MaxRetry`
  - Maximum retry count when the handler returns an error.

- `RetryDelay`
  - Delay before retry.

- `BufferSize`
  - Only relevant for in-memory queue implementations.

A logical queue must be registered before handlers are attached to it. Re-registering the same queue requires an identical config.

## Available Queue Backends

### Memory Queue

Package:

```text
infrastructure/asynctask/queue/memoryqueue
```

Characteristics:

- in-process only
- simple and fast
- good for local development or single-process deployment

Limitations:

- no cross-process durability
- tasks are lost if the process exits

### RabbitMQ Queue

Package:

```text
infrastructure/asynctask/queue/rabbitmq
```

Characteristics:

- durable broker-backed queue
- suitable for multi-process consumers
- supports ack and retry through republish

Good for:

- production queue workloads
- durable background jobs
- workloads that need broker-side isolation

### Redis Stream Queue

Package:

```text
infrastructure/asynctask/queue/redisstream
```

Characteristics:

- built on Redis Stream + Consumer Group
- supports ack, retry, and pending reclaim
- lighter operationally if Redis already exists in the system

Good for:

- moderate queue workloads
- systems already using Redis heavily
- scenarios where introducing RabbitMQ is unnecessary

## Factory Registration

Factories live in `infrastructure/asynctask/factory`.

### In-Memory Queue

`Ants` and `Thunnus` factories only register task-pool-based infrastructure:

- one `TaskDispatcher`
- one `CronTaskDispatcher`

Example:

```go
builders := scene.ModuleFactoryArray{
	logger.ZapFactory{}.Default(),
	asynctask.Ants{},
}
```

If you also need queue support, register a queue backend separately, for example:

```go
builders := scene.ModuleFactoryArray{
	logger.ZapFactory{}.Default(),
	asynctask.Ants{},
	asynctask.MemoryQueue{},
}
```

### RabbitMQ Queue

Use:

```go
builders := scene.ModuleFactoryArray{
	logger.ZapFactory{}.Default(),
	asynctask.Ants{},
	asynctask.RabbitMQ{}.Default(),
}
```

Config keys:

- `rabbitmq.url`
- `rabbitmq.exchange`

Optional defaults are already provided for:

- `ExchangeType = direct`
- `Durable = true`
- `Prefetch = 16`

### Redis Stream Queue

Use:

```go
builders := scene.ModuleFactoryArray{
	logger.ZapFactory{}.Default(),
	asynctask.Ants{},
	asynctask.RedisStream{}.Default(),
}
```

Config keys:

- `redis.host`
- `redis.port`
- `redis.username`
- `redis.password`
- `redis.database`
- `asynctask.redisstream.stream_prefix`
- `asynctask.redisstream.group_prefix`

## Registry Usage

Queue interfaces are registered through the normal registry API.

```go
registry.Register[asynctask.TaskQueuePublisher](taskQueue)
registry.Register[asynctask.TaskQueueConsumer](taskQueue)
```

Business code should inject the queue interfaces, not the concrete backend.

Example:

```go
type myService struct {
	queuePub asynctask.TaskQueuePublisher `aperture:""`
	queueCon asynctask.TaskQueueConsumer  `aperture:""`
}
```

## Typical Usage Pattern

There are two sides:

1. register queue config on startup
2. register handlers on startup
3. publish tasks during business execution

When publisher and worker live in different processes, they should not do the same thing:

- publisher
  - only injects `TaskQueuePublisher`
  - only publishes tasks

- worker
  - injects `TaskQueueConsumer`
  - registers queue config and handlers
  - owns the consumption lifecycle

In the current scene architecture, a worker is best modeled as a `void` scene application.
That means:

- the publisher usually lives in an HTTP/RPC-facing scene
- the worker usually lives in a `void.VoidApp`
- queue registration is typically done in the worker app's `Run()` after services are initialized

Example:

```go
type CacheWorker struct {
	queueCon asynctask.TaskQueueConsumer `aperture:""`
}

func (w *CacheWorker) Run() error {
	if err := w.queueCon.RegisterQueue("meowsic.media-cache", asynctask.TaskQueueConfig{
		Concurrency: 2,
		MaxRetry:    3,
		RetryDelay:  3 * time.Second,
	}); err != nil {
		return err
	}

	return w.queueCon.RegisterHandler(
		"meowsic.media-cache",
		"meowsic.media-cache.file",
		asynctask.TaskQueueHandlerFunc(w.handleCacheFile),
	)
}
```

This matches the current `void` contract:

- service initialization finishes first
- delivery layer starts after that
- `void` app `Run()` starts background delivery behavior and returns quickly

### 1. Register a queue and then register a handler

Usually do this in `Setup()` of a service.

```go
type myTaskPayload struct {
	UserID string `json:"user_id"`
}

type myService struct {
	queueCon asynctask.TaskQueueConsumer `aperture:""`
}

func (s *myService) Setup() error {
	if err := s.queueCon.RegisterQueue("user.sync", asynctask.TaskQueueConfig{
		Concurrency: 2,
		MaxRetry:    3,
		RetryDelay:  5 * time.Second,
	}); err != nil {
		return err
	}

	return s.queueCon.RegisterHandler(
		"user.sync",
		"user.sync.profile",
		asynctask.TaskQueueHandlerFunc(func(ctx context.Context, task *asynctask.QueueTask) error {
			var payload myTaskPayload
			if err := asynctask.UnmarshalPayload(task, &payload); err != nil {
				return err
			}
			return s.handleUserSync(ctx, payload.UserID)
		}),
	)
}
```

### 2. Publish a task

```go
func (s *myService) EnqueueUserSync(userID string) error {
	payload, err := asynctask.MarshalPayload(myTaskPayload{
		UserID: userID,
	})
	if err != nil {
		return err
	}

	_, err = s.queuePub.Publish(context.Background(), &asynctask.QueueTask{
		Queue:    "user.sync",
		Type:     "user.sync.profile",
		Payload:  payload,
		Timeout:  30 * time.Second,
		MaxRetry: 3,
	})
	return err
}
```

## Payload Helpers

Use helper functions to encode and decode JSON payloads.

### Encode

```go
payload, err := asynctask.MarshalPayload(myStruct)
```

### Decode

```go
var payload myStruct
err := asynctask.UnmarshalPayload(task, &payload)
```

## Queue Naming Recommendation

Keep `Queue` and `Type` separate.

Recommended pattern:

- `Queue`
  - represents resource isolation and concurrency domain
  - examples:
    - `meowsic.media-cache`
    - `notify.send`
    - `user.sync`

- `Type`
  - represents business action inside the queue
  - examples:
    - `meowsic.media-cache.file`
    - `meowsic.media-cache.info`
    - `notify.send.email`

This separation is important:

- queue controls worker count and retry policy
- task type controls handler routing

## Retry Semantics

When `HandleTask` returns a non-nil error:

- the queue backend marks the task as retrying
- if `Attempt < MaxRetry`, the task is requeued
- if retry count is exhausted, the task becomes failed

Current behavior:

- memory queue requeues in-process
- RabbitMQ republishes the task
- Redis Stream republishes the task into the stream and acks the old message

## Delay Semantics

`QueueTask.Delay` is backend-dependent.

Current behavior:

- memory queue supports delay directly
- RabbitMQ implementation stores the requested delay in task metadata but does not provide broker-native delayed delivery
- Redis Stream implementation republishes and waits until `AvailableAt`

If strict delayed delivery is required in production, prefer:

- RabbitMQ with delayed-message support or TTL/DLX strategy
- Redis with a dedicated delayed-task design

## Concurrency Semantics

Concurrency is limited per logical queue, not globally.

Example:

- `meowsic.media-cache` with `Concurrency: 2`
- `notify.send` with `Concurrency: 16`

This means cache jobs will not occupy all generic background capacity.

That is the primary reason to use the queue abstraction instead of `TaskDispatcher`.

## Cron Integration

`CronTaskDispatcher` still uses `TaskDispatcher`, not the queue abstraction.

This is intentional:

- cron is a trigger mechanism
- `TaskDispatcher` remains the simplest local execution model

If a cron job needs queue isolation, the cron callback should publish a queue task explicitly.

Example:

```go
cron.Add("0 */5 * * * *", func() error {
	_, err := queuePub.Publish(context.Background(), &asynctask.QueueTask{
		Queue: "report.sync",
		Type:  "report.sync.daily",
	})
	return err
})
```

## Recommended Design Rules

- Inject interfaces, not concrete backends.
- Keep `Payload` small.
- Put large objects into storage and pass only IDs or references.
- Use one logical queue for one concurrency domain.
- Use multiple task types within the same queue only when they should share concurrency limits.
- Register handlers during service setup, not lazily during request handling.
- Do not put transport-specific DTOs directly into queue payloads if domain-oriented payloads are cleaner.

## Choosing a Backend

Choose `memoryqueue` when:

- local development
- single process
- no durability needed

Choose `rabbitmq` when:

- durable queue is required
- multiple consumers/processes are required
- explicit broker-backed tasking is desired

Choose `redisstream` when:

- Redis already exists in the system
- moderate reliability is enough
- you want a queue without running RabbitMQ

## Example Module Wiring

```go
builders := scene.ModuleFactoryArray{
	logger.ZapFactory{}.Default(),
	asynctask.Ants{},
	asynctask.RedisStream{}.Default(),
}
```

Then in the service:

```go
type svc struct {
	queuePub asynctask.TaskQueuePublisher `aperture:""`
	queueCon asynctask.TaskQueueConsumer  `aperture:""`
}
```

And in `Setup()`:

```go
func (s *svc) Setup() error {
	if err := s.queueCon.RegisterQueue("meowsic.media-cache", asynctask.TaskQueueConfig{
		Concurrency: 2,
		MaxRetry:    3,
		RetryDelay:  3 * time.Second,
	}); err != nil {
		return err
	}

	return s.queueCon.RegisterHandler(
		"meowsic.media-cache",
		"meowsic.media-cache.file",
		asynctask.TaskQueueHandlerFunc(s.handleCacheTask),
	)
}
```

## Current Limitations

- queue task status is tracked in-memory in the task object; there is no built-in task status repository yet
- no built-in dead-letter queue abstraction yet
- delayed execution semantics are not equally strong across all backends
- RabbitMQ and Redis Stream implementations are functional first versions and can still be extended for richer production features

## Task Status Limitation

`Publish` only tells the caller whether the task was accepted by the queue backend.
It does not provide a reliable way to query the final execution result.

This is especially important when:

- publisher and worker are in different processes
- the backend is RabbitMQ or Redis Stream
- tasks may retry or be consumed by another instance

Today, `QueueTask.Status` is only a runtime field on the task object itself.
It is not backed by a shared task status store.

That means:

- publisher can know whether publish succeeded
- publisher cannot reliably know whether the task eventually succeeded or failed
- there is currently no built-in `GetTaskStatus(id)` style API

For `memoryqueue`, the publisher may appear to observe status changes when it still holds the same in-memory `*QueueTask` pointer.
This is only an implementation side effect of sharing the same process memory.
It is not a portable or stable task status mechanism, and business code should not depend on it.

If final task status is required, the recommended next step is to add a dedicated task status repository or let business handlers persist their own result state explicitly.

## Next Recommended Enhancements

- dead-letter queue abstraction
- task status persistence and query API
- metrics hooks for publish, success, retry and failure
- backend-specific integration tests
- delayed queue support with stronger semantics
