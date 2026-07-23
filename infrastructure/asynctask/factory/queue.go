package factory

import (
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	queueimpl "github.com/rhine-tech/scene/infrastructure/asynctask/queue"
	"github.com/rhine-tech/scene/infrastructure/asynctask/queue/asynq"
	"github.com/rhine-tech/scene/infrastructure/asynctask/queue/rabbitmq"
	"github.com/rhine-tech/scene/infrastructure/asynctask/queue/redisstream"
	"github.com/rhine-tech/scene/infrastructure/asynctask/reporter/memoryreporter"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/registry"
)

type MemoryQueue struct {
	scene.ModuleFactory
}

func (b MemoryQueue) Init() scene.LensInit {
	return func() {
		taskQueue := registry.Load(queueimpl.NewMemoryTaskQueue())
		registry.Register[asynctask.TaskQueuePublisher](taskQueue)
		registry.Register[asynctask.TaskQueueConsumer](taskQueue)
	}
}

type MemoryReporter struct {
	scene.ModuleFactory
	Config memoryreporter.Config
}

func (b MemoryReporter) Init() scene.LensInit {
	return func() {
		reporter := memoryreporter.NewMemoryReporter(b.Config)
		registry.Register[asynctask.TaskReporter](reporter)
		registry.Register[asynctask.TaskQueueInspector](reporter)
	}
}

func (b MemoryReporter) Default() MemoryReporter {
	return MemoryReporter{
		Config: memoryreporter.Config{
			EventTTL:        time.Duration(registry.Config.GetInt("asynctask.memoryreporter.event_ttl_seconds")) * time.Second,
			MaxTasks:        int(registry.Config.GetInt("asynctask.memoryreporter.max_tasks")),
			CleanupInterval: time.Duration(registry.Config.GetInt("asynctask.memoryreporter.cleanup_interval_seconds")) * time.Second,
		},
	}
}

type RabbitMQ struct {
	scene.ModuleFactory
	Config rabbitmq.Config
}

func (b RabbitMQ) Init() scene.LensInit {
	return func() {
		taskQueue := registry.Load(rabbitmq.New(b.Config))
		registry.Register[asynctask.TaskQueuePublisher](taskQueue)
		registry.Register[asynctask.TaskQueueConsumer](taskQueue)
	}
}

func (b RabbitMQ) Default() RabbitMQ {
	return RabbitMQ{
		Config: rabbitmq.Config{
			URL:          registry.Config.GetString("rabbitmq.url"),
			Exchange:     registry.Config.GetString("rabbitmq.exchange"),
			ExchangeType: "direct",
			Durable:      true,
			Prefetch:     16,
		},
	}
}

type Asynq struct {
	scene.ModuleFactory
	Config asynq.Config
}

func (b Asynq) Init() scene.LensInit {
	return func() {
		taskQueue := registry.Load(asynq.New(b.Config))
		registry.Register[asynctask.TaskQueuePublisher](taskQueue)
		registry.Register[asynctask.TaskQueueConsumer](taskQueue)
	}
}

func (b Asynq) Default() Asynq {
	return Asynq{
		Config: asynq.Config{
			Redis: datasource.RedisConfig{
				Host:     registry.Config.GetString("redis.host"),
				Port:     int(registry.Config.GetInt("redis.port")),
				Username: registry.Config.GetString("redis.username"),
				Password: registry.Config.GetString("redis.password"),
				Database: int(registry.Config.GetInt("redis.database")),
			},
		},
	}
}

type RedisStream struct {
	scene.ModuleFactory
	Config redisstream.Config
}

func (b RedisStream) Init() scene.LensInit {
	return func() {
		taskQueue := registry.Load(redisstream.New(b.Config))
		registry.Register[asynctask.TaskQueuePublisher](taskQueue)
		registry.Register[asynctask.TaskQueueConsumer](taskQueue)
	}
}

func (b RedisStream) Default() RedisStream {
	return RedisStream{
		Config: redisstream.Config{
			Redis: datasource.RedisConfig{
				Host:     registry.Config.GetString("redis.host"),
				Port:     int(registry.Config.GetInt("redis.port")),
				Username: registry.Config.GetString("redis.username"),
				Password: registry.Config.GetString("redis.password"),
				Database: int(registry.Config.GetInt("redis.database")),
			},
			StreamPrefix: registry.Config.GetString("asynctask.redisstream.stream_prefix"),
			GroupPrefix:  registry.Config.GetString("asynctask.redisstream.group_prefix"),
		},
	}
}
