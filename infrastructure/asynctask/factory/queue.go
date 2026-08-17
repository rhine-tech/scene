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
	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/registry"
)

type MemoryQueue struct {
	scene.ModuleFactory
}

func (b MemoryQueue) Init(container *registry.Container) {
	taskQueue := queueimpl.NewMemoryTaskQueue()
	registry.Export[asynctask.TaskQueuePublisher](container, taskQueue)
	registry.Export[asynctask.TaskQueueConsumer](container, taskQueue)
}

type MemoryReporter struct {
	scene.ModuleFactory
	Config memoryreporter.Config
}

func (b MemoryReporter) Init(container *registry.Container) {
	reporter := memoryreporter.NewMemoryReporter(b.Config)
	registry.Export[asynctask.TaskReporter](container, reporter)
	registry.Export[asynctask.TaskQueueInspector](container, reporter)
}

func (b MemoryReporter) Default() MemoryReporter {
	cfg := registry.Use[config.IConfig](nil)
	return MemoryReporter{
		Config: memoryreporter.Config{
			EventTTL:        time.Duration(cfg.GetInt("asynctask.memoryreporter.event_ttl_seconds")) * time.Second,
			MaxTasks:        int(cfg.GetInt("asynctask.memoryreporter.max_tasks")),
			CleanupInterval: time.Duration(cfg.GetInt("asynctask.memoryreporter.cleanup_interval_seconds")) * time.Second,
		},
	}
}

type RabbitMQ struct {
	scene.ModuleFactory
	Config rabbitmq.Config
}

func (b RabbitMQ) Init(container *registry.Container) {
	taskQueue := rabbitmq.New(b.Config)
	registry.Export[asynctask.TaskQueuePublisher](container, taskQueue)
	registry.Export[asynctask.TaskQueueConsumer](container, taskQueue)
}

func (b RabbitMQ) Default() RabbitMQ {
	cfg := registry.Use[config.IConfig](nil)
	return RabbitMQ{
		Config: rabbitmq.Config{
			URL:          cfg.GetString("rabbitmq.url"),
			Exchange:     cfg.GetString("rabbitmq.exchange"),
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

func (b Asynq) Init(container *registry.Container) {
	taskQueue := asynq.New(b.Config)
	registry.Export[asynctask.TaskQueuePublisher](container, taskQueue)
	registry.Export[asynctask.TaskQueueConsumer](container, taskQueue)
}

func (b Asynq) Default() Asynq {
	cfg := registry.Use[config.IConfig](nil)
	return Asynq{
		Config: asynq.Config{
			Redis: datasource.RedisConfig{
				Host:     cfg.GetString("redis.host"),
				Port:     int(cfg.GetInt("redis.port")),
				Username: cfg.GetString("redis.username"),
				Password: cfg.GetString("redis.password"),
				Database: int(cfg.GetInt("redis.database")),
			},
		},
	}
}

type RedisStream struct {
	scene.ModuleFactory
	Config redisstream.Config
}

func (b RedisStream) Init(container *registry.Container) {
	taskQueue := redisstream.New(b.Config)
	registry.Export[asynctask.TaskQueuePublisher](container, taskQueue)
	registry.Export[asynctask.TaskQueueConsumer](container, taskQueue)
}

func (b RedisStream) Default() RedisStream {
	cfg := registry.Use[config.IConfig](nil)
	return RedisStream{
		Config: redisstream.Config{
			Redis: datasource.RedisConfig{
				Host:     cfg.GetString("redis.host"),
				Port:     int(cfg.GetInt("redis.port")),
				Username: cfg.GetString("redis.username"),
				Password: cfg.GetString("redis.password"),
				Database: int(cfg.GetInt("redis.database")),
			},
			StreamPrefix: cfg.GetString("asynctask.redisstream.stream_prefix"),
			GroupPrefix:  cfg.GetString("asynctask.redisstream.group_prefix"),
		},
	}
}
