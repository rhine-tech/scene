package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type graphA interface{ GraphA() }
type graphB interface{ GraphB() }

type injectedGraphAImpl struct {
	b graphB `aperture:""`
}

type injectedGraphBImpl struct {
	a graphA `aperture:""`
}

func (*injectedGraphAImpl) GraphA() {}
func (*injectedGraphBImpl) GraphB() {}

func TestDependencyOrder(t *testing.T) {
	tests := []struct {
		name         string
		dependencies [][]int
		want         []int
	}{
		{name: "chain", dependencies: [][]int{{1}, {2}, nil}, want: []int{2, 1, 0}},
		{name: "independent nodes", dependencies: [][]int{nil, nil, nil, nil}, want: []int{0, 1, 2, 3}},
		{name: "stable ready order", dependencies: [][]int{{2}, nil, nil}, want: []int{1, 2, 0}},
		{name: "newly ready declaration priority", dependencies: [][]int{{2}, {3}, nil, nil}, want: []int{2, 0, 3, 1}},
		{name: "diamond", dependencies: [][]int{{1, 2}, {3}, {3}, nil}, want: []int{3, 1, 2, 0}},
		{name: "diamond reversed edges", dependencies: [][]int{{2, 1}, {3}, {3}, nil}, want: []int{3, 1, 2, 0}},
		{name: "cycle", dependencies: [][]int{{1}, {0}, {0}}, want: []int{0, 1, 2}},
		{name: "independent cycles", dependencies: [][]int{{1}, {0}, {3}, {2}, {0, 2}}, want: []int{0, 1, 2, 3, 4}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, dependencyOrder(test.dependencies))
		})
	}
}

func TestScopeAllowsInjectionCycleAcrossContainers(t *testing.T) {
	firstValue := new(injectedGraphAImpl)
	first := NewContainer("first")
	Export[graphA](first, firstValue)

	secondValue := new(injectedGraphBImpl)
	second := NewContainer("second")
	Export[graphB](second, secondValue)

	scope := NewScope()
	require.NoError(t, scope.Build(first, second))
	require.Same(t, secondValue, firstValue.b)
	require.Same(t, firstValue, secondValue.a)
	require.Equal(t, []any{firstValue, secondValue}, scope.OrderedValues())
}

func TestScopeOwnsOneModuleSet(t *testing.T) {
	scope := NewScope()
	global := &testDependencyImpl{value: "global"}
	SetIn[testDependency](scope, global)

	first := NewContainer("first")
	firstValue := &testDependencyImpl{value: "first"}
	Export[testDependency](first, firstValue)
	require.NoError(t, scope.Build(first))
	require.Same(t, firstValue, UseIn[testDependency](scope, nil))

	second := NewContainer("second")
	secondValue := &testDependencyImpl{value: "second"}
	Export[testDependency](second, secondValue)
	require.ErrorContains(t, scope.Build(second), "already contains module dependencies")

	scope.Reset()
	require.Empty(t, scope.OrderedValues())
	require.Same(t, global, UseIn[testDependency](scope, nil))
	require.NoError(t, scope.Build(second))
	require.Same(t, secondValue, UseIn[testDependency](scope, nil))
}
