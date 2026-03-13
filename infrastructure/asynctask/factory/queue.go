package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	queueimpl "github.com/rhine-tech/scene/infrastructure/asynctask/queue"
	"github.com/rhine-tech/scene/infrastructure/asynctask/queue/asynq"
	"github.com/rhine-tech/scene/infrastructure/asynctask/queue/rabbitmq"
	"github.com/rhine-tech/scene/infrastructure/asynctask/queue/redisstream"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/registry"
)

type MemoryQueue struct {
	scene.ModuleFactory
}

func (b MemoryQueue) Init() scene.LensInit {
	return func() {
		taskQueue := queueimpl.NewMemoryTaskQueue()
		registry.Register[asynctask.TaskQueuePublisher](taskQueue)
		registry.Register[asynctask.TaskQueueConsumer](taskQueue)
	}
}

type RabbitMQ struct {
	scene.ModuleFactory
	Config rabbitmq.Config
}

func (b RabbitMQ) Init() scene.LensInit {
	return func() {
		taskQueue := rabbitmq.New(b.Config)
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
		taskQueue := asynq.New(b.Config)
		registry.Register[asynctask.TaskQueuePublisher](taskQueue)
		registry.Register[asynctask.TaskQueueConsumer](taskQueue)
	}
}

func (b Asynq) Default() Asynq {
	return Asynq{
		Config: asynq.Config{
			Redis: datasource.DatabaseConfig{
				Host:     registry.Config.GetString("redis.host"),
				Port:     int(registry.Config.GetInt("redis.port")),
				Username: registry.Config.GetString("redis.username"),
				Password: registry.Config.GetString("redis.password"),
				Database: registry.Config.GetString("redis.database"),
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
		taskQueue := redisstream.New(b.Config)
		registry.Register[asynctask.TaskQueuePublisher](taskQueue)
		registry.Register[asynctask.TaskQueueConsumer](taskQueue)
	}
}

func (b RedisStream) Default() RedisStream {
	return RedisStream{
		Config: redisstream.Config{
			Redis: datasource.DatabaseConfig{
				Host:     registry.Config.GetString("redis.host"),
				Port:     int(registry.Config.GetInt("redis.port")),
				Username: registry.Config.GetString("redis.username"),
				Password: registry.Config.GetString("redis.password"),
				Database: registry.Config.GetString("redis.database"),
			},
			StreamPrefix: registry.Config.GetString("asynctask.redisstream.stream_prefix"),
			GroupPrefix:  registry.Config.GetString("asynctask.redisstream.group_prefix"),
		},
	}
}
