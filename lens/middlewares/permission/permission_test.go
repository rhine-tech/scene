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
