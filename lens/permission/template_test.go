package permission

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplateBind(t *testing.T) {
	template := MustParseTemplate("Project:{Project_ID}:file:{file_id}:read")
	perm, err := template.Bind("ABC-123", "File_2")
	require.NoError(t, err)
	require.Equal(t, "project:abc-123:file:file_2:read", perm.String())

	other := template.MustBind("other", "file_2")
	require.Equal(t, "project:other:file:file_2:read", other.String())
	require.Equal(t, "project:abc-123:file:file_2:read", perm.String())
	require.Equal(t, "project:{project_id}:file:{file_id}:read", template.String())

	grant := MustParsePermission("project:abc-123")
	require.True(t, grant.HasPermission(perm))
	require.False(t, grant.HasPermission(other))
}

func TestTemplateBindRejectsInvalidArguments(t *testing.T) {
	template := MustParseTemplate("project:{project_id}:read")
	for _, args := range [][]string{
		nil, {"123", "456"}, {""}, {"123:run"}, {"{other_id}"},
		{"a/b"}, {"a.b"}, {"a b"}, {"*"}, {"%"}, {"项目"},
	} {
		perm, err := template.Bind(args...)
		require.Error(t, err, args)
		require.Nil(t, perm, args)
	}
	require.Panics(t, func() { template.MustBind("123:run") })
}

func TestParseTemplate(t *testing.T) {
	t.Run("static path", func(t *testing.T) {
		template, err := ParseTemplate("Project:Read")
		require.NoError(t, err)
		require.Equal(t, "project:read", template.MustBind().String())
	})

	t.Run("invalid paths", func(t *testing.T) {
		for _, path := range []string{
			"", "project::read", "project:{}:read", "project:{id:read",
			"project:id}:read", "project:prefix-{id}:read", "project:{{id}}:read",
			"project:{project.id}:read", "project:{id}:read*",
		} {
			template, err := ParseTemplate(path)
			require.Error(t, err, path)
			require.Nil(t, template, path)
		}
	})
}
