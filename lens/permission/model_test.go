package permission

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// --- Permission Struct Tests ---

func TestPermission_ParseAndString(t *testing.T) {
	t.Run("valid multi-level permission", func(t *testing.T) {
		p, err := ParsePermission("user:edit:profile")
		require.NoError(t, err)
		require.NotNil(t, p)
		require.Equal(t, []string{"user", "edit", "profile"}, p.parts)
		// Test the String() method for reconstruction
		require.Equal(t, "user:edit:profile", p.String())
	})

	t.Run("valid single-level permission", func(t *testing.T) {
		p, err := ParsePermission("admin")
		require.NoError(t, err)
		require.NotNil(t, p)
		require.Equal(t, []string{"admin"}, p.parts)
		require.Equal(t, "admin", p.String())
	})

	t.Run("normalizes uppercase permission", func(t *testing.T) {
		p, err := ParsePermission("Project:Member_01:READ-ONLY")
		require.NoError(t, err)
		require.Equal(t, "project:member_01:read-only", p.String())
	})

	t.Run("error cases for ParsePermission", func(t *testing.T) {
		for _, name := range []string{
			"", "user::edit", "user:edit:", ":user",
			"project:read write", "project:read\n", "project:file/read",
			"project:file.read", "project:*", "project:123%", "project:123!",
			"project:{project_id}:read", "project:项目:read",
		} {
			p, err := ParsePermission(name)
			require.Error(t, err, name)
			require.Nil(t, p, name)
		}
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
	perm1.parts[2] = "d"
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
	subPerm.parts[0] = "x"
	require.Equal(t, "a:b:c:d", perm1.String())
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

	t.Run("rejects unbound templates", func(t *testing.T) {
		var perm Permission
		require.Error(t, json.Unmarshal([]byte(`"project:{project_id}:read"`), &perm))
	})
}

// --- DeclarationTree Tests ---

func TestDeclarationTree_ToList(t *testing.T) {
	t.Run("empty tree", func(t *testing.T) {
		tree := NewDeclarationTree()
		list := tree.ToList()
		require.Empty(t, list, "ToList on an empty tree should return an empty slice")
	})

	t.Run("single permission", func(t *testing.T) {
		tree := NewDeclarationTree()
		tree.Add(MustParsePermission("admin"))
		list := tree.ToList()
		require.Len(t, list, 1)
		require.Equal(t, "admin", list[0])
	})

	t.Run("multiple simple permissions", func(t *testing.T) {
		tree := NewDeclarationTree()
		tree.Add(
			MustParsePermission("admin"),
			MustParsePermission("user"),
			MustParsePermission("guest"),
		)
		expected := []string{"admin", "guest", "user"}
		require.ElementsMatch(t, expected, tree.ToList())
	})

	t.Run("complex tree with overlapping paths", func(t *testing.T) {
		tree := NewDeclarationTree()
		tree.Add(
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
		require.ElementsMatch(t, expected, tree.ToList())
	})

	t.Run("tree where a prefix is also a terminal node", func(t *testing.T) {
		tree := NewDeclarationTree()
		tree.Add(
			MustParsePermission("user:edit"),
			MustParsePermission("user:edit:profile"),
		)
		expected := []string{"user:edit", "user:edit:profile"}
		require.ElementsMatch(t, expected, tree.ToList())
	})
}
