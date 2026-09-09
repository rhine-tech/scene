package permission

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Permission struct {
	parts []string
}

// OwnerPermissions groups the explicit grants selected for an owner.
type OwnerPermissions struct {
	Owner       string        `json:"owner"`
	Permissions []*Permission `json:"permissions"`
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

func MustParsePermission(name string) *Permission {
	p, err := ParsePermission(name)
	if err != nil {
		panic(err)
	}
	return p
}

// ParsePermission parses colon-separated, non-empty permission segments.
// Segments may contain ASCII letters, digits, underscores and hyphens.
// ASCII letters are normalized to lowercase.
func ParsePermission(name string) (*Permission, error) {
	perms := strings.Split(name, ":")
	for i, perm := range perms {
		var err error
		perms[i], err = normalizePermissionPart(perm)
		if err != nil {
			return nil, err
		}
	}
	return &Permission{parts: perms}, nil
}

func normalizePermissionPart(part string) (string, error) {
	if part == "" {
		return "", errors.New("permission: contains zero length permission string")
	}
	for i := 0; i < len(part); i++ {
		c := part[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-' {
			continue
		}
		return "", fmt.Errorf("permission: invalid segment %q: only ASCII letters, digits, '_' and '-' are allowed", part)
	}
	return strings.ToLower(part), nil
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
