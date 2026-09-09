package permission

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeclarationTreeList(t *testing.T) {
	tree := NewDeclarationTree()
	for _, path := range []string{
		"project:{project_id}",
		"project:{project_id}:run",
		"project:{project_id}:run:read",
		"project:{project_id}:run:read:own",
		"project:{project_id}:run:create",
		"project:{project_id}:function:invoke",
	} {
		tree.AddTemplate(MustParseTemplate(path))
	}
	tree.Add(MustParsePermission("project:fixed:run:create"))

	prefix := MustParseTemplate("project:{project_id}:run")
	templates := tree.List(prefix)
	paths := make([]string, len(templates))
	bound := make([]string, len(templates))
	for i, template := range templates {
		paths[i] = template.String()
		bound[i] = template.MustBind("123").String()
	}
	require.ElementsMatch(t, []string{
		"project:{project_id}:run",
		"project:{project_id}:run:read",
		"project:{project_id}:run:read:own",
		"project:{project_id}:run:create",
	}, paths)
	require.ElementsMatch(t, []string{
		"project:123:run",
		"project:123:run:read",
		"project:123:run:read:own",
		"project:123:run:create",
	}, bound)

	// Structural prefixes can select declarations without being grants themselves.
	functions := tree.List(MustParseTemplate("project:{project_id}:function"))
	require.Len(t, functions, 1)
	require.Equal(t, "project:{project_id}:function:invoke", functions[0].String())
	require.Empty(t, tree.List(MustParseTemplate("project:{project_id}:revision")))
	require.Empty(t, tree.List(MustParseTemplate("project:123:run")))

	concrete := tree.List(MustParseTemplate("project:fixed:run"))
	require.Len(t, concrete, 1)
	require.Equal(t, "project:fixed:run:create", concrete[0].MustBind().String())
}
