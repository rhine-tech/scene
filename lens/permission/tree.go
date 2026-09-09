package permission

import "strings"

// PermissionNode represents a node in the permission Trie.
type PermissionNode struct {
	// IsTerminal marks the end of a valid permission.
	// For the permission "user:edit", the node "edit" would have IsTerminal = true.
	IsTerminal bool `json:"is_terminal"`
	// IsPlaceholder marks a placeholder segment such as "{project_id}".
	// It does not mark literal descendants of a placeholder.
	IsPlaceholder bool `json:"is_placeholder"`
	// Children holds the next parts of the permission.
	// e.g., if this node is "user", a child might be "edit".
	Children map[string]*PermissionNode `json:"children"`
}

func NewPermissionNode() *PermissionNode {
	return &PermissionNode{
		Children: make(map[string]*PermissionNode),
	}
}

func (node *PermissionNode) add(parts []string) {
	current := node
	for _, part := range parts {
		child, exists := current.Children[part]
		if !exists {
			child = NewPermissionNode()
			child.IsPlaceholder = strings.HasPrefix(part, "{")
			current.Children[part] = child
		}
		current = child
	}
	current.IsTerminal = true // Mark the last node as the end of a permission
}

func (node *PermissionNode) walk(path []string, visit func([]string)) {
	if node.IsTerminal {
		visit(path)
	}
	for part, child := range node.Children {
		child.walk(append(path, part), visit)
	}
}

// DeclarationTree describes available permissions and templates for display.
// Terminal nodes mark declarations, not grants. Placeholders remain named
// path segments; this tree does not perform permission checks.
type DeclarationTree struct {
	Root *PermissionNode
}

// NewDeclarationTree creates an empty permission declaration tree.
func NewDeclarationTree() *DeclarationTree {
	return &DeclarationTree{Root: NewPermissionNode()}
}

// Add declares concrete permissions in the tree.
func (d *DeclarationTree) Add(perms ...*Permission) {
	for _, perm := range perms {
		if perm != nil {
			d.Root.add(perm.parts)
		}
	}
}

// AddTemplate declares templates without binding their placeholders.
func (d *DeclarationTree) AddTemplate(templates ...*Template) {
	for _, template := range templates {
		d.Root.add(template.parts)
	}
}

// ToList returns declared paths, including unbound template placeholders.
func (d *DeclarationTree) ToList() []string {
	var paths []string
	d.Root.walk(nil, func(path []string) {
		paths = append(paths, strings.Join(path, ":"))
	})
	return paths
}
