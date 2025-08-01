package permission

// RootPermTree hold all permission created in the application
var RootPermTree *PermissionTree = NewPermissionTree()

// Create calls MustParsePermission, but it also adds this permission
// to RootPermTree
func Create(name string) *Permission {
	p := MustParsePermission(name)
	RootPermTree.Add(p)
	return p
}
