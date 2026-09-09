package permission

import (
	"fmt"
	"strings"
)

// Template is a permission path with whole-segment placeholders, such as
// "project:{project_id}:file:{file_id}:read". A template must be bound before
// it can be used as a grant or checked for access.
// Templates are immutable and may be bound concurrently.
type Template struct {
	parts      []string
	parameters []int
}

// ParseTemplate parses a permission template. Literal segments and placeholder
// names follow the same character and lowercase rules as ParsePermission.
// Each placeholder occurrence consumes one argument, from left to right.
func ParseTemplate(name string) (*Template, error) {
	parts := strings.Split(name, ":")
	var parameters []int
	for i, part := range parts {
		placeholder := strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}")
		if placeholder {
			part = part[1 : len(part)-1]
		}
		normalized, err := normalizePermissionPart(part)
		if err != nil {
			return nil, err
		}
		if placeholder {
			parts[i] = "{" + normalized + "}"
			parameters = append(parameters, i)
		} else {
			parts[i] = normalized
		}
	}
	return &Template{parts: parts, parameters: parameters}, nil
}

// MustParseTemplate is ParseTemplate but panics if the template is invalid.
func MustParseTemplate(name string) *Template {
	template, err := ParseTemplate(name)
	if err != nil {
		panic(err)
	}
	return template
}

// String returns the normalized path, including its named placeholders.
func (t *Template) String() string {
	return strings.Join(t.parts, ":")
}

// Bind returns a concrete permission by filling placeholders from left to right.
// The argument count must match the number of placeholders. Each argument must
// be a single non-empty segment and follows ParsePermission's character rules.
// Binding does not register or grant the resulting permission.
func (t *Template) Bind(args ...string) (*Permission, error) {
	if len(args) != len(t.parameters) {
		return nil, fmt.Errorf("permission: template expects %d arguments, got %d", len(t.parameters), len(args))
	}
	parts := make([]string, len(t.parts))
	copy(parts, t.parts)
	for i, position := range t.parameters {
		value, err := normalizePermissionPart(args[i])
		if err != nil {
			return nil, fmt.Errorf("permission: argument %d for %s: %w", i+1, t.parts[position], err)
		}
		parts[position] = value
	}
	return &Permission{parts: parts}, nil
}

// MustBind is Bind but panics if the arguments are invalid.
func (t *Template) MustBind(args ...string) *Permission {
	perm, err := t.Bind(args...)
	if err != nil {
		panic(err)
	}
	return perm
}

// Match extracts positional arguments from a concrete permission whose full
// path matches the template. It does not check permission inheritance.
// Arguments follow Bind's left-to-right order; a mismatch returns nil, false.
func (t *Template) Match(perm *Permission) (args []string, matched bool) {
	if perm == nil || len(t.parts) != len(perm.parts) {
		return nil, false
	}
	args = make([]string, len(t.parameters))
	parameter := 0
	for i, part := range t.parts {
		if parameter < len(t.parameters) && t.parameters[parameter] == i {
			args[parameter] = perm.parts[i]
			parameter++
		} else if part != perm.parts[i] {
			return nil, false
		}
	}
	return args, true
}
