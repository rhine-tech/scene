package registry

import (
	"slices"
)

func appendUniqueInt(values []int, value int) []int {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

// dependencyOrder topologically orders strongly connected components while
// preserving container declaration order inside each component.
func dependencyOrder(dependencies [][]int) []int {
	indices := make([]int, len(dependencies))
	lowlinks := make([]int, len(dependencies))
	onStack := make([]bool, len(dependencies))
	for index := range indices {
		indices[index] = -1
	}
	stack := make([]int, 0, len(dependencies))
	components := make([][]int, 0)
	nextIndex := 0
	var connect func(int)
	connect = func(node int) {
		indices[node] = nextIndex
		lowlinks[node] = nextIndex
		nextIndex++
		stack = append(stack, node)
		onStack[node] = true
		for _, dependency := range dependencies[node] {
			if indices[dependency] == -1 {
				connect(dependency)
				lowlinks[node] = min(lowlinks[node], lowlinks[dependency])
			} else if onStack[dependency] {
				lowlinks[node] = min(lowlinks[node], indices[dependency])
			}
		}
		if lowlinks[node] != indices[node] {
			return
		}
		component := make([]int, 0)
		for {
			last := len(stack) - 1
			member := stack[last]
			stack = stack[:last]
			onStack[member] = false
			component = append(component, member)
			if member == node {
				break
			}
		}
		slices.Sort(component)
		components = append(components, component)
	}
	for node := range dependencies {
		if indices[node] == -1 {
			connect(node)
		}
	}

	componentOf := make([]int, len(dependencies))
	for componentIndex, component := range components {
		for _, node := range component {
			componentOf[node] = componentIndex
		}
	}
	componentDependencies := make([][]int, len(components))
	dependents := make([][]int, len(components))
	for node, nodeDependencies := range dependencies {
		consumer := componentOf[node]
		for _, dependency := range nodeDependencies {
			provider := componentOf[dependency]
			if consumer == provider {
				continue
			}
			componentDependencies[consumer] = appendUniqueInt(componentDependencies[consumer], provider)
			dependents[provider] = appendUniqueInt(dependents[provider], consumer)
		}
	}

	remainingDependencies := make([]int, len(components))
	for componentIndex := range components {
		remainingDependencies[componentIndex] = len(componentDependencies[componentIndex])
	}
	processed := make([]bool, len(components))
	order := make([]int, 0, len(dependencies))
	for len(order) < len(dependencies) {
		selected := -1
		selectedFirst := len(dependencies)
		for componentIndex, component := range components {
			if processed[componentIndex] || remainingDependencies[componentIndex] != 0 {
				continue
			}
			if component[0] < selectedFirst {
				selected = componentIndex
				selectedFirst = component[0]
			}
		}
		processed[selected] = true
		order = append(order, components[selected]...)
		for _, consumer := range dependents[selected] {
			remainingDependencies[consumer]--
		}
	}
	return order
}
