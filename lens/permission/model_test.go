package permission

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

// --- Permission Struct Tests ---

func TestPermission_ParseAndString(t *testing.T) {
	t.Run("valid multi-level permission", func(t *testing.T) {
		p, err := ParsePermission("user:edit:profile")
		require.NoError(t, err)
		require.NotNil(t, p)
		require.Equal(t, "user", p.Name)
		require.NotNil(t, p.SubPermission)
		require.Equal(t, "edit", p.SubPermission.Name)
		require.NotNil(t, p.SubPermission.SubPermission)
		require.Equal(t, "profile", p.SubPermission.SubPermission.Name)
		require.Nil(t, p.SubPermission.SubPermission.SubPermission)
		// Test the String() method for reconstruction
		require.Equal(t, "user:edit:profile", p.String())
	})

	t.Run("valid single-level permission", func(t *testing.T) {
		p, err := ParsePermission("admin")
		require.NoError(t, err)
		require.NotNil(t, p)
		require.Equal(t, "admin", p.Name)
		require.Nil(t, p.SubPermission)
		require.Equal(t, "admin", p.String())
	})

	t.Run("error cases for ParsePermission", func(t *testing.T) {
		_, err := ParsePermission("")
		require.Error(t, err, "should fail on empty string")

		_, err = ParsePermission("user::edit")
		require.Error(t, err, "should fail on empty part")

		_, err = ParsePermission("user:edit:")
		require.Error(t, err, "should fail on trailing colon")
	})

	t.Run("MustParsePermission panics on error", func(t *testing.T) {
		require.Panics(t, func() {
			MustParsePermission("a::b")
		})
		require.NotPanics(t, func() {
			MustParsePermission("a:b")
		})
	})
}

func TestPermission_IsEqual(t *testing.T) {
	p1 := MustParsePermission("a:b:c")
	p2 := MustParsePermission("a:b:c")
	p3 := MustParsePermission("a:b:d")
	p4 := MustParsePermission("a:b")

	require.True(t, p1.IsEqual(p2), "p1 should be equal to p2")
	require.False(t, p1.IsEqual(p3), "p1 should not be equal to p3 (different sub)")
	require.False(t, p1.IsEqual(p4), "p1 should not be equal to p4 (p4 is shorter)")
	require.False(t, p4.IsEqual(p1), "p4 should not be equal to p1 (p1 is longer)")
	require.False(t, p1.IsEqual(nil), "any permission should not be equal to nil")
}

func TestPermission_HasPermission(t *testing.T) {
	// A user who has "user:edit" permission
	userEditPerm := MustParsePermission("user:edit")
	// A user who has top-level "admin" permission
	adminPerm := MustParsePermission("admin")

	// Exact match
	require.True(t, userEditPerm.HasPermission(MustParsePermission("user:edit")))
	// Sub-permission check: user:edit grants access to user:edit:profile
	require.True(t, userEditPerm.HasPermission(MustParsePermission("user:edit:profile")))
	// Wildcard check: admin grants access to anything starting with admin
	require.True(t, adminPerm.HasPermission(MustParsePermission("admin:users:delete")))
	require.True(t, adminPerm.HasPermission(MustParsePermission("admin:dashboard")))

	// A more specific permission does NOT grant a more general one
	require.False(t, userEditPerm.HasPermission(MustParsePermission("user")))
	// A sibling permission is not granted
	require.False(t, userEditPerm.HasPermission(MustParsePermission("user:view")))
	// A different root permission is not granted
	require.False(t, userEditPerm.HasPermission(MustParsePermission("admin:edit")))
}

func TestPermission_Copy(t *testing.T) {
	perm := MustParsePermission("a:b:c")
	perm1 := perm.Copy()

	// They should be equal right after copy
	require.True(t, perm.IsEqual(perm1))
	// The pointers should be different
	require.NotSame(t, perm, perm1)

	// Ensure it's a deep copy by modifying the copy
	perm1.SubPermission.SubPermission.Name = "d"
	require.False(t, perm.IsEqual(perm1), "modifying copy should not affect original")
	require.Equal(t, "a:b:c", perm.String())
	require.Equal(t, "a:b:d", perm1.String())
}

func TestPermission_WithSubPerm(t *testing.T) {
	perm := MustParsePermission("a:b")
	subPerm := MustParsePermission("c:d")

	// Chain a parsed sub-permission
	perm1 := perm.WithSubPerm(subPerm)
	require.Equal(t, "a:b:c:d", perm1.String())

	// Original should be unchanged
	require.Equal(t, "a:b", perm.String())

	// Sub-permission should be a copy
	require.NotSame(t, subPerm, perm1.SubPermission.SubPermission)
}

func TestPermission_WithSubPermStr(t *testing.T) {
	perm := MustParsePermission("a:b:c")
	perm1 := perm.WithSubPermStr("d")
	perm2 := perm.WithSubPermStr("d").WithSubPermStr("e")
	perm3 := perm.WithSubPermStr("d:e:f") // Test multi-part string

	require.Equal(t, "a:b:c:d", perm1.String())
	require.Equal(t, "a:b:c:d:e", perm2.String())
	require.Equal(t, "a:b:c:d:e:f", perm3.String())
	// Original should be unchanged
	require.Equal(t, "a:b:c", perm.String())
}

func TestPermission_JSONMarshaling(t *testing.T) {
	t.Run("single permission", func(t *testing.T) {
		perm := MustParsePermission("a:b:c")
		val, err := json.Marshal(perm)
		require.NoError(t, err)
		require.Equal(t, `"a:b:c"`, string(val))

		var p2 Permission
		err = json.Unmarshal(val, &p2)
		require.NoError(t, err)
		require.True(t, perm.IsEqual(&p2))
	})

	t.Run("slice of permissions", func(t *testing.T) {
		perms := []*Permission{
			MustParsePermission("a:b"),
			MustParsePermission("x:y:z"),
		}
		val, err := json.Marshal(perms)
		require.NoError(t, err)
		require.Equal(t, `["a:b","x:y:z"]`, string(val))

		var perms2 []*Permission
		err = json.Unmarshal(val, &perms2)
		require.NoError(t, err)
		require.Len(t, perms2, 2)
		require.True(t, perms[0].IsEqual(perms2[0]))
		require.True(t, perms[1].IsEqual(perms2[1]))
	})
}

// --- PermissionTree (Trie) Tests ---

func TestPermissionTree_AddAndHasPermission(t *testing.T) {
	tree := NewPermissionTree()
	tree.Add(
		MustParsePermission("user:view"),
		MustParsePermission("report:generate:pdf"),
		MustParsePermission("admin"), // Wildcard admin
	)

	// --- Test Cases That Should Pass ---
	// Exact matches
	require.True(t, tree.HasPermission(MustParsePermission("user:view")), "exact match failed")
	require.True(t, tree.HasPermissionStr("report:generate:pdf"), "exact match str failed")

	// Prefix/wildcard matches
	require.True(t, tree.HasPermissionStr("admin:dashboard:view"), "wildcard admin failed")
	require.True(t, tree.HasPermissionStr("user:view:book"), "wildcard user:view failed")
	require.True(t, tree.HasPermissionStr("admin:users:delete"), "wildcard admin failed")
	require.True(t, tree.HasPermission(MustParsePermission("admin")), "exact wildcard match failed")

	// --- Test Cases That Should Fail ---
	// No match
	require.False(t, tree.HasPermissionStr("user:delete"), "mismatched action should fail")
	require.False(t, tree.HasPermissionStr("report:delete"), "mismatched sibling should fail")

	// Specific permission does not grant general
	require.False(t, tree.HasPermissionStr("user"), "specific should not grant general")
	require.False(t, tree.HasPermissionStr("report:generate"), "specific should not grant general")

	// Invalid string
	require.False(t, tree.HasPermissionStr("user::view"), "invalid string should fail safely")
}

func TestPermissionTree_EdgeCases(t *testing.T) {
	t.Run("empty tree", func(t *testing.T) {
		tree := NewPermissionTree()
		require.False(t, tree.HasPermissionStr("any:perm"))
	})

	t.Run("adding parent perm after child", func(t *testing.T) {
		tree := NewPermissionTree()
		tree.Add(MustParsePermission("a:b:c"))
		// Before adding parent, check should fail
		require.False(t, tree.HasPermissionStr("a:b:d"))

		// Now add the parent, which acts as a wildcard for this level
		tree.Add(MustParsePermission("a:b"))
		require.True(t, tree.HasPermissionStr("a:b:d"), "check should pass after parent was added")
		require.True(t, tree.HasPermissionStr("a:b:c"), "original perm should still pass")
	})

	t.Run("adding child perm after parent", func(t *testing.T) {
		tree := NewPermissionTree()
		tree.Add(MustParsePermission("a:b"))
		require.True(t, tree.HasPermissionStr("a:b:c")) // Has access due to parent

		// Add a more specific one (should have no negative effect)
		tree.Add(MustParsePermission("a:b:c"))
		require.True(t, tree.HasPermissionStr("a:b:c"))
		require.True(t, tree.HasPermissionStr("a:b:d"))
	})

	t.Run("adding duplicate permissions", func(t *testing.T) {
		tree := NewPermissionTree()
		tree.Add(MustParsePermission("a:b"))
		tree.Add(MustParsePermission("a:b"))
		tree.Add(MustParsePermission("a:b"))
		require.True(t, tree.HasPermissionStr("a:b:c"))
		// A simple structural check
		require.Len(t, tree.Root.Children, 1)
		require.Len(t, tree.Root.Children["a"].Children, 1)
	})

	t.Run("check against nil permission", func(t *testing.T) {
		tree := NewPermissionTree()
		tree.Add(MustParsePermission("a"))
		require.False(t, tree.HasPermission(nil))
	})
}

func permsToStrings(perms []*Permission) []string {
	strs := make([]string, len(perms))
	for i, p := range perms {
		strs[i] = p.String()
	}
	sort.Strings(strs)
	return strs
}

func TestPermissionTree_ToList(t *testing.T) {
	t.Run("empty tree", func(t *testing.T) {
		tree := NewPermissionTree()
		list := tree.ToList()
		require.Empty(t, list, "ToList on an empty tree should return an empty slice")
	})

	t.Run("single permission", func(t *testing.T) {
		tree := BuildTree(MustParsePermission("admin"))
		list := tree.ToList()
		require.Len(t, list, 1)
		require.Equal(t, "admin", list[0].String())
	})

	t.Run("multiple simple permissions", func(t *testing.T) {
		tree := BuildTree(
			MustParsePermission("admin"),
			MustParsePermission("user"),
			MustParsePermission("guest"),
		)
		expected := []string{"admin", "guest", "user"}
		actual := permsToStrings(tree.ToList())
		require.Equal(t, expected, actual)
	})

	t.Run("complex tree with overlapping paths", func(t *testing.T) {
		tree := BuildTree(
			MustParsePermission("user:view"),
			MustParsePermission("user:edit:profile"),
			MustParsePermission("user:edit:avatar"),
			MustParsePermission("admin"),
			MustParsePermission("report:generate"),
			MustParsePermission("report:view"),
		)
		// Add a duplicate to ensure it's handled correctly
		tree.Add(MustParsePermission("admin"))

		expected := []string{
			"admin",
			"report:generate",
			"report:view",
			"user:edit:avatar",
			"user:edit:profile",
			"user:view",
		}
		actual := permsToStrings(tree.ToList())
		require.Equal(t, expected, actual)
	})

	t.Run("tree where a prefix is also a terminal node", func(t *testing.T) {
		tree := BuildTree(
			MustParsePermission("user:edit"),
			MustParsePermission("user:edit:profile"),
		)
		expected := []string{"user:edit", "user:edit:profile"}
		actual := permsToStrings(tree.ToList())
		require.Equal(t, expected, actual)
	})
}
