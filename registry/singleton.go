package registry

import (
	"fmt"
	"sync"
)

var singletonRegistry = make(map[string]interface{})
var singletonLock sync.RWMutex

func RegisterSingleton[T any](singleton interface{}) {
	RegisterSingletonByName(getInterfaceName[T](), singleton)
}

func RegisterSingletonByName(name string, singleton interface{}) {
	singletonLock.Lock()
	singletonRegistry[name] = singleton
	singletonLock.Unlock()
}

func AcquireSingleton[T any](iface T) T {
	singletonLock.RLock()
	tpName := getInterfaceName[T]()
	impl, ok := singletonRegistry[tpName]
	singletonLock.RUnlock()
	if !ok {
		panic(fmt.Sprintf("no singleton registered for %s", tpName))
	}
	return impl.(T)
}

func AcquireSingletonByName[T any](name string) T {
	singletonLock.RLock()
	impl, ok := singletonRegistry[name]
	singletonLock.RUnlock()
	if !ok {
		panic(fmt.Sprintf("no singleton registered for %s", name))
	}
	return impl.(T)
}
