package registry

import (
	"fmt"
)

// Build validates, connects, and injects a flat set of module containers.
func (s *Scope) Build(containers ...*Container) error {
	s.lock.RLock()
	built := s.providers != nil
	s.lock.RUnlock()
	if built {
		return fmt.Errorf("scene registry: scope already contains module dependencies")
	}

	moduleContainers := append([]*Container(nil), containers...)
	providers := make(map[string]int)
	for index, container := range moduleContainers {
		if container == nil {
			return fmt.Errorf("scene registry: module container %d is nil", index)
		}
		container.sealed = true
		for _, name := range container.Provides() {
			if existingIndex, ok := providers[name]; ok {
				return fmt.Errorf(
					"scene registry: dependency %s is exported by both %s and %s",
					name,
					moduleContainers[existingIndex].name,
					container.name,
				)
			}
			providers[name] = index
		}
	}

	dependencies := make([][]int, len(containers))
	for index, container := range moduleContainers {
		for _, name := range container.Requires(false) {
			providerIndex, ok := providers[name]
			if !ok {
				return fmt.Errorf("scene registry: module %s requires missing dependency %s", container.name, name)
			}
			dependencies[index] = appendUniqueInt(dependencies[index], providerIndex)
		}
		for _, name := range container.Requires(true) {
			providerIndex, ok := providers[name]
			if !ok {
				continue
			}
			dependencies[index] = appendUniqueInt(dependencies[index], providerIndex)
		}
	}
	order := dependencyOrder(dependencies)
	for _, container := range moduleContainers {
		names := append(container.Requires(false), container.Requires(true)...)
		for _, name := range names {
			sourceIndex, exists := providers[name]
			if !exists {
				continue
			}
			source := moduleContainers[sourceIndex]
			value, ok := source.Resolve(name)
			if !ok {
				return fmt.Errorf("scene registry: dependency %s from %s is unavailable", name, source.name)
			}
			container.Import(name, value)
		}
	}
	for _, index := range order {
		moduleContainers[index].Inject(s.runInjectHooks)
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.providers != nil {
		return fmt.Errorf("scene registry: scope already contains module dependencies")
	}
	s.containers = moduleContainers
	s.providers = providers
	s.order = order
	return nil
}

// Reset removes the built module dependencies while preserving global
// values and injection hooks so this Scope can build another module set.
func (s *Scope) Reset() {
	s.lock.Lock()
	s.containers = nil
	s.providers = nil
	s.order = nil
	s.lock.Unlock()
}

// OrderedValues returns module-owned values in dependency order and in each
// container's declaration order.
func (s *Scope) OrderedValues() []any {
	s.lock.RLock()
	defer s.lock.RUnlock()
	values := make([]any, 0)
	for _, index := range s.order {
		values = append(values, s.containers[index].values()...)
	}
	return values
}
