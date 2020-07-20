package perms

import (
	"strings"
)

// Module permission.
type Permission string

func (p Permission) String() string {
	return string(p)
}

// Slice of Permission objects.
type Permissions []Permission

func (list Permissions) String() string {
	out := make([]string, 0, len(list))
	for _, p := range list {
		out = append(out, p.String())
	}

	return strings.Join(out, ", ")
}

// RequestModulePermissions handler that returns requested perms for module.
type RequestModulePermissions func() (moduleName string, modulePermissions Permissions)
