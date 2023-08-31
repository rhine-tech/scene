package permission

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPermission_MarshalJSON(t *testing.T) {
	perm := MustParsePermission("a.b.c")
	perm1 := MustParsePermission("a.b.c.d")
	pset := PermissionSet([]*Permission{perm, perm1})
	val, err := json.Marshal(pset)
	assert.Nil(t, err)
	var pset2 PermissionSet
	err = json.Unmarshal(val, &pset2)
	assert.Nil(t, err)
	fmt.Println(pset[0], pset2[0])
	assert.True(t, pset[0].IsEqual(pset2[0]))
	assert.Equal(t, `["a.b.c","a.b.c.d"]`, string(val))
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

	x, _ := json.MarshalIndent(BuildTreeFromSet(perms), "", "  ")
	fmt.Println(string(x))
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
