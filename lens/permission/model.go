package permission

import (
	"encoding/json"
	"errors"
	"strings"
)

type PermOwner string

func (p PermOwner) String() string {
	return string(p)
}

type Permission struct {
	parts []string
}

func (p *Permission) Parts() []string {
	if p == nil {
		return nil
	}
	out := make([]string, len(p.parts))
	copy(out, p.parts)
	return out
}

func (p *Permission) UnmarshalJSON(bytes []byte) error {
	var val string
	if err := json.Unmarshal(bytes, &val); err != nil {
		return err
	}
	p1, err := ParsePermission(val)
	if err != nil {
		return err
	}
	*p = *p1
	return nil
}

func (p *Permission) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func newPermission(perms []string) *Permission {
	if len(perms) == 0 {
		return nil
	}
	copied := make([]string, len(perms))
	copy(copied, perms)
	return &Permission{parts: copied}
}

func MustParsePermission(name string) *Permission {
	p, err := ParsePermission(name)
	if err != nil {
		panic(err)
	}
	return p
}

func ParsePermission(name string) (*Permission, error) {
	perms := strings.Split(name, ":")
	if len(perms) == 0 {
		return nil, errors.New("permission: contains zero length permission string")
	}
	for _, perm := range perms {
		if len(perm) == 0 {
			return nil, errors.New("permission: contains zero length permission string")
		}
	}
	return newPermission(perms), nil
}

func (p *Permission) String() string {
	if p == nil {
		return ""
	}
	return strings.Join(p.parts, ":")
}

func (p *Permission) HasPermission(perm *Permission) bool {
	if p == nil || perm == nil {
		return false
	}
	if len(p.parts) == 0 || len(perm.parts) == 0 {
		return false
	}
	if p.parts[0] != perm.parts[0] {
		return false
	}
	if len(p.parts) == 1 {
		return true
	}
	if len(perm.parts) == 1 {
		return false
	}
	if len(p.parts) > len(perm.parts) {
		return false
	}
	for i := 1; i < len(p.parts); i++ {
		if p.parts[i] != perm.parts[i] {
			return false
		}
	}
	return true
}

func (p *Permission) IsEqual(perm *Permission) bool {
	if (p == nil) && (perm == nil) {
		return false
	}
	if (p == nil) || (perm == nil) {
		return false
	}
	if len(p.parts) != len(perm.parts) {
		return false
	}
	for i := range p.parts {
		if p.parts[i] != perm.parts[i] {
			return false
		}
	}
	return true
}

func (p *Permission) copy(last *Permission) *Permission {
	if p == nil {
		return nil
	}
	parts := append([]string{}, p.parts...)
	if last != nil {
		parts = append(parts, last.parts...)
	}
	return &Permission{parts: parts}
}

func (p *Permission) Copy() *Permission {
	return p.copy(nil)
}

// WithSubPerm returns a new permission with sub permission
func (p *Permission) WithSubPerm(perm *Permission) *Permission {
	if perm == nil {
		return p.Copy()
	}
	return p.copy(perm.Copy())
}

// WithSubPermStr returns a new permission with sub permission
// - perm must be a valid permission string
func (p *Permission) WithSubPermStr(perm string) *Permission {
	return p.copy(MustParsePermission(perm))
}

// PermissionNode represents a node in the permission Trie.
type PermissionNode struct {
	// IsTerminal marks the end of a valid permission.
	// For the permission "user:edit", the node "edit" would have IsTerminal = true.
	IsTerminal bool `json:"is_terminal"`
	// Children holds the next parts of the permission.
	// e.g., if this node is "user", a child might be "edit".
	Children map[string]*PermissionNode `json:"children"`
}

func NewPermissionNode() *PermissionNode {
	return &PermissionNode{
		Children: make(map[string]*PermissionNode),
	}
}

// PermissionTree is the main structure that holds all permissions for fast lookups.
type PermissionTree struct {
	Root *PermissionNode
}

// NewPermissionTree creates an empty permission tree.
func NewPermissionTree() *PermissionTree {
	return &PermissionTree{
		Root: NewPermissionNode(),
	}
}

// BuildTree creates tree from permissions
func BuildTree(permissions ...*Permission) *PermissionTree {
	tree := &PermissionTree{
		Root: NewPermissionNode(),
	}
	tree.Add(permissions...)
	return tree
}

// add is internal method inserts a permission into the Trie.
func (pt *PermissionTree) add(perm *Permission) {
	if perm == nil || len(perm.parts) == 0 {
		return
	}
	node := pt.Root
	for i := 0; i < len(perm.parts); i++ {
		part := perm.parts[i]
		child, exists := node.Children[part]
		if !exists {
			child = NewPermissionNode()
			node.Children[part] = child
		}
		node = child
	}
	node.IsTerminal = true // Mark the last node as the end of a permission
}

func (pt *PermissionTree) Add(perms ...*Permission) {
	for _, perm := range perms {
		pt.add(perm)
	}
}

// HasPermission checks if the tree grants the given permission.
// This is the high-speed replacement for PermissionSet.HasPermission.
// Complexity: O(D), where D is the depth of the permission being checked.
func (pt *PermissionTree) HasPermission(perm *Permission) bool {
	node := pt.Root
	if perm == nil || len(perm.parts) == 0 {
		return false
	}
	for _, part := range perm.parts {
		// Check for a "wildcard" permission. If a parent path is a valid
		// permission in the tree, access is granted.
		// e.g., if the tree has "user", this check will return true for "user:edit".
		if node.IsTerminal {
			return true
		}

		child, exists := node.Children[part]
		if !exists {
			// The path does not exist in the tree.
			return false
		}

		node = child
	}

	// The full path was found. We must check if this final node or any of its
	// ancestors was a terminal node.
	return node.IsTerminal
}

// HasPermissionStr is a helper to check using a string, similar to your original API
func (pt *PermissionTree) HasPermissionStr(perm string) bool {
	p, err := ParsePermission(perm)
	if err != nil {
		return false // Or handle error appropriately
	}
	return pt.HasPermission(p)
}

// ToList returns all terminal nodes in the permission tree.
func (pt *PermissionTree) ToList() []*Permission {
	var perms []*Permission
	var path []string
	var dfs func(*PermissionNode)
	dfs = func(node *PermissionNode) {
		if node.IsTerminal {
			perms = append(perms, newPermission(path))
		}
		for part, child := range node.Children {
			path = append(path, part)
			dfs(child)
			path = path[:len(path)-1]
		}
	}
	dfs(pt.Root)
	return perms
}
