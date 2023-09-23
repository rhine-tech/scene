package permission

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPermission_MarshalJSON(t *testing.T) {
	perm := MustParsePermission("a:b:c")
	perm1 := MustParsePermission("a:b:c:d")
	pset := PermissionSet([]*Permission{perm, perm1})
	val, err := json.Marshal(pset)
	require.Nil(t, err)
	var pset2 PermissionSet
	err = json.Unmarshal(val, &pset2)
	require.Nil(t, err)
	require.True(t, pset[0].IsEqual(pset2[0]))
	require.Equal(t, `["a:b:c","a:b:c:d"]`, string(val))
}

func TestPermission_Tree(t *testing.T) {
	perms := PermissionSet{
		MustParsePermission("a:b:c"),
		MustParsePermission("a:b:d:d"),
		MustParsePermission("a:b:c:e"),
		MustParsePermission("b:d:c:f"),
		MustParsePermission("b:d:e:f"),
		MustParsePermission("r:s:t"),
	}

	_, err := json.MarshalIndent(BuildTreeFromSet(perms), "", "  ")
	require.NoError(t, err)
}

func TestPermissionSet_Cleanup(t *testing.T) {
	perms := PermissionSet{
		MustParsePermission("a:b:c"),
		MustParsePermission("a:b:d:d"),
		MustParsePermission("a:b:c:e"),
		MustParsePermission("b:d:c:f"),
		MustParsePermission("b:d:e:f"),
		MustParsePermission("b:d:e"),
		MustParsePermission("r:s:t"),
	}
	fmt.Println(perms.Cleanup())
}

func TestPermission_Copy(t *testing.T) {
	perm := MustParsePermission("a:b:c")
	perm1 := perm.Copy()
	require.True(t, perm.IsEqual(perm1))
	perm1.SubPermission.SubPermission.Name = "d"
	require.False(t, perm.IsEqual(perm1))
}
