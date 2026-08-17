package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"reflect"
)

func LoggerAddPrefix() registry.InjectHook {
	return registry.InjectHook{
		Interface: "logger.ILogger",
		Hook: func(iface string, obj reflect.Value, fieldVal reflect.Value, i *interface{}) {
			value := obj.Interface()
			name := ""
			if s, ok := value.(scene.Named); ok {
				name = s.ImplName().Identifier()
			}
			if name != "" {
				*i = (*i).(logger.ILogger).WithPrefix(name)
			}
		},
	}
}
