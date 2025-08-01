package permission

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type PermOwner string

func (p PermOwner) String() string {
	return string(p)
}

type Permission struct {
	Name          string      `json:",ommitempty"`
	SubPermission *Permission `json:",ommitempty"`
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

func NewPermission(perms []string) *Permission {
	if len(perms) == 0 {
		return nil
	}
	var subperm *Permission
	if len(perms) > 1 {
		subperm = NewPermission(perms[1:])
	}
	if len(perms[0]) == 0 {
		panic(errors.New("permission: contains zero length permission string"))
	}
	return &Permission{
		Name:          perms[0],
		SubPermission: subperm,
	}
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
	return NewPermission(perms), nil
}

func (p *Permission) String() string {
	if p.SubPermission != nil {
		return fmt.Sprintf("%s:%s", p.Name, p.SubPermission.String())
	}
	return p.Name
}

func (p *Permission) HasPermission(perm *Permission) bool {
	if p.Name != perm.Name {
		return false
	}
	// if no sub permission, means this permission is the top level permission
	if p.SubPermission == nil {
		return true
	}
	// if sub permission is nil, target permission is top level permission.
	// so it must be false
	if perm.SubPermission == nil {
		return false
	}
	return p.SubPermission.HasPermission(perm.SubPermission)
}

func (p *Permission) IsEqual(perm *Permission) bool {
	if (p == nil) && (perm == nil) {
		return false
	}
	if (p == nil) || (perm == nil) {
		return false
	}
	if p.Name != perm.Name {
		return false
	}
	if p.SubPermission == nil && perm.SubPermission == nil {
		return true
	}
	if p.SubPermission == nil || perm.SubPermission == nil {
		return false
	}
	return p.SubPermission.IsEqual(perm.SubPermission)
}

func (p *Permission) copy(last *Permission) *Permission {
	if p == nil {
		return nil
	}
	// can be implemented by recursion or ParsePermission
	// but since it is tail recursion, so I use loop
	// Base case for the recursion

	// top level permission
	newPerm := &Permission{Name: "", SubPermission: &Permission{}}
	// holder is used to hold the previous permission
	prevPerm := newPerm
	currPerm := p
	for currPerm != nil {
		prevPerm = prevPerm.SubPermission
		prevPerm.Name = currPerm.Name
		currPerm = currPerm.SubPermission
		prevPerm.SubPermission = &Permission{}
	}
	// last permission should be nil
	prevPerm.SubPermission = last
	return newPerm.SubPermission
}

func (p *Permission) Copy() *Permission {
	return p.copy(nil)
}

// WithSubPerm returns a new permission with sub permission
func (p *Permission) WithSubPerm(perm *Permission) *Permission {
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
	// Name is the full name of this permission node
	Name string `json:"name"`
	// Children holds the next parts of the permission.
	// e.g., if this node is "user", a child might be "edit".
	Children map[string]*PermissionNode `json:"children"`
}

func NewPermissionNode() *PermissionNode {
	return &PermissionNode{
		Children: make(map[string]*PermissionNode),
	}
}

// PermTree is the main structure that holds all permissions for fast lookups.
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
	node := pt.Root
	p := perm
	prefix := ""
	for p != nil {
		child, exists := node.Children[p.Name]
		if !exists {
			child = NewPermissionNode()
			child.Name = prefix + p.Name
			node.Children[p.Name] = child
		}
		prefix = child.Name + ":"
		node = child
		p = p.SubPermission
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
	p := perm
	for p != nil {
		// Check for a "wildcard" permission. If a parent path is a valid
		// permission in the tree, access is granted.
		// e.g., if the tree has "user", this check will return true for "user:edit".
		if node.IsTerminal {
			return true
		}

		child, exists := node.Children[p.Name]
		if !exists {
			// The path does not exist in the tree.
			return false
		}

		node = child
		p = p.SubPermission
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
	var dfs func(*PermissionNode)
	dfs = func(node *PermissionNode) {
		if node.IsTerminal {
			// As the full permission string is stored in the node's Name,
			// we can parse it directly.
			p, err := ParsePermission(node.Name)
			if err == nil { // Should not fail if names are valid
				perms = append(perms, p)
			}
		}
		for _, child := range node.Children {
			dfs(child)
		}
	}
	dfs(pt.Root)
	return perms
}

//// PermissionSet Depracated
//// BuildTreeFromSet takes a slice of permissions and builds the efficient Trie structure.
//func BuildTreeFromSet(perms PermissionSet) *PermissionTree {
//	tree := NewPermissionTree()
//	tree.Add(perms...)
//	return tree
//}
//
//
//
//type PermissionSet []*Permission
//
//func (ps PermissionSet) HasPermission(perm *Permission) bool {
//	for _, p1 := range ps {
//		if p1.HasPermission(perm) {
//			return true
//		}
//	}
//	return false
//}
//
//func (ps PermissionSet) HasPermissionStr(perm string) bool {
//	pm, err := ParsePermission(perm)
//	if err != nil {
//		return false
//	}
//	for _, p1 := range ps {
//		if p1.HasPermission(pm) {
//			return true
//		}
//	}
//	return false
//}
//
//// Merge merges two permission set
//func (ps PermissionSet) Merge(other PermissionSet) PermissionSet {
//	x := mapset.NewSet[*Permission](ps...).Union(
//		mapset.NewSet[*Permission](other...))
//	return x.ToSlice()
//}
//
//// Cleanup cleans up the permission set
//// - if permission set already have a top level permission, then remove all sub permissions
//func (ps PermissionSet) Cleanup() PermissionSet {
//	topPerm := make(map[string]*Permission)
//	childMap := make(map[string]PermissionSet)
//	for _, perm := range ps {
//		permName := perm.Name
//		if perm.SubPermission == nil {
//			topPerm[permName] = perm
//			continue
//		}
//		childMap[permName] = append(childMap[permName], perm.SubPermission)
//	}
//	var newPs PermissionSet
//	for _, perm := range topPerm {
//		newPs = append(newPs, &Permission{
//			Name: perm.Name,
//		})
//	}
//	for key, perms := range childMap {
//		if _, ok := topPerm[key]; ok {
//			continue
//		}
//		var subPerms PermissionSet
//		for _, perm := range perms.Cleanup() {
//			subPerms = append(subPerms, &Permission{
//				Name:          key,
//				SubPermission: perm,
//			})
//		}
//		newPs = append(newPs, subPerms...)
//	}
//	return newPs
//}
//
//func (ps PermissionSet) ToStrSlice() []string {
//	retval := make([]string, len(ps))
//	for i, p := range ps {
//		retval[i] = p.String()
//	}
//	return retval
//}
