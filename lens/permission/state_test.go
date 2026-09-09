package permission

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPermissionDeclarationsAndBoundGrants(t *testing.T) {
	original := RootPermTree
	RootPermTree = NewDeclarationTree()
	t.Cleanup(func() { RootPermTree = original })

	Create("project")
	project := CreateTemplate("project:{project_id}")
	read := CreateTemplate("project:{project_id}:read")
	CreateTemplate("project:{project_id}:read")
	expected := []string{"project", "project:{project_id}", "project:{project_id}:read"}
	require.ElementsMatch(t, expected, RootPermTree.ToList())

	data, err := json.Marshal(RootPermTree.Root.Children)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"project": {"is_terminal": true, "is_placeholder": false, "children": {
			"{project_id}": {"is_terminal": true, "is_placeholder": true, "children": {
				"read": {"is_terminal": true, "is_placeholder": false, "children": {}}
			}}
		}}
	}`, string(data))

	grant := project.MustBind("123")
	require.True(t, grant.HasPermission(read.MustBind("123")))
	require.False(t, grant.HasPermission(read.MustBind("456")))
	require.Equal(t, "project:123", grant.String())
	require.ElementsMatch(t, expected, RootPermTree.ToList())
}
