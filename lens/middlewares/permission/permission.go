package permission

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
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

type PermissionSet []*Permission

func (ps PermissionSet) HasPermission(perm *Permission) bool {
	for _, p1 := range ps {
		if p1.HasPermission(perm) {
			return true
		}
	}
	return false
}

func (ps PermissionSet) HasPermissionStr(perm string) bool {
	pm, err := ParsePermission(perm)
	if err != nil {
		return false
	}
	for _, p1 := range ps {
		if p1.HasPermission(pm) {
			return true
		}
	}
	return false
}

func (ps PermissionSet) Merge(other PermissionSet) PermissionSet {
	x := mapset.NewSet[*Permission](ps...).Union(
		mapset.NewSet[*Permission](other...))
	return x.ToSlice()
}

func (ps PermissionSet) Cleanup() PermissionSet {
	topPerm := make(map[string]*Permission)
	childMap := make(map[string]PermissionSet)
	for _, perm := range ps {
		permName := perm.Name
		if perm.SubPermission == nil {
			topPerm[permName] = perm
			continue
		}
		childMap[permName] = append(childMap[permName], perm.SubPermission)
	}
	var newPs PermissionSet
	for _, perm := range topPerm {
		newPs = append(newPs, &Permission{
			Name: perm.Name,
		})
	}
	for key, perms := range childMap {
		if _, ok := topPerm[key]; ok {
			continue
		}
		var subPerms PermissionSet
		for _, perm := range perms.Cleanup() {
			subPerms = append(subPerms, &Permission{
				Name:          key,
				SubPermission: perm,
			})
		}
		newPs = append(newPs, subPerms...)
	}
	return newPs
}

func (ps PermissionSet) ToStrSlice() []string {
	retval := make([]string, len(ps))
	for i, p := range ps {
		retval[i] = p.String()
	}
	return retval
}

type PermTree map[string]PermTree

func BuildTreeFromSet(ps PermissionSet) PermTree { // Root node to hold everything
	return buildChildren("", ps)
}

func buildChildren(parent string, ps PermissionSet) PermTree {
	if len(ps) == 0 {
		return nil
	}
	tree := make(PermTree)
	childMap := make(map[string][]*Permission)

	for _, perm := range ps {
		permName := perm.Name
		if _, ok := childMap[permName]; !ok {
			childMap[permName] = nil
		}
		if perm.SubPermission == nil {
			continue
		}
		childMap[permName] = append(childMap[permName], perm.SubPermission)
	}

	for key, childs := range childMap {
		currentKey := parent + key
		tree[currentKey] = buildChildren(currentKey+":", childs)
	}
	return tree
}
