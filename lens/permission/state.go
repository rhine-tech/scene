package permission

// RootPermTree holds the permissions and templates declared in the application.
// Populate it during initialization, before serving requests.
var RootPermTree *DeclarationTree = NewDeclarationTree()

// Create calls MustParsePermission, but it also adds this permission
// to RootPermTree
func Create(name string) *Permission {
	p := MustParsePermission(name)
	RootPermTree.Add(p)
	return p
}

// CreateTemplate calls MustParseTemplate and adds the template to RootPermTree.
// Declare templates during initialization, alongside permissions created by Create.
func CreateTemplate(name string) *Template {
	template := MustParseTemplate(name)
	RootPermTree.AddTemplate(template)
	return template
}
