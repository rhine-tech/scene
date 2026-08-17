package registry

import "sync"

// Scope owns the global values, injection hooks, and module dependencies
// for one isolated registry environment.
type Scope struct {
	lock       sync.RWMutex
	globals    map[string]any
	hooks      map[string][]InjectHookFunc
	containers []*Container
	providers  map[string]int
	order      []int
}

func NewScope() *Scope {
	return &Scope{
		globals: make(map[string]any),
		hooks:   make(map[string][]InjectHookFunc),
	}
}

var defaultScope = NewScope()

// DefaultScope returns the Scope used by package-level registry functions.
func DefaultScope() *Scope {
	return defaultScope
}

func (s *Scope) set(name string, value any) {
	s.lock.Lock()
	s.globals[name] = value
	s.lock.Unlock()
}

func (s *Scope) lookup(name string) (any, bool) {
	s.lock.RLock()
	if index, ok := s.providers[name]; ok {
		value, exists := s.containers[index].Resolve(name)
		s.lock.RUnlock()
		return value, exists
	}
	value, exists := s.globals[name]
	s.lock.RUnlock()
	return value, exists
}

func (s *Scope) injectionHooks(name string) []InjectHookFunc {
	s.lock.RLock()
	hooks := append([]InjectHookFunc(nil), s.hooks[name]...)
	s.lock.RUnlock()
	return hooks
}
