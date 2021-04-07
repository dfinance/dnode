package perms

import (
	"fmt"
	"strings"

	"github.com/dfinance/dnode/helpers"
)

// ModulePermissions holds module available permission and per other module cross permissions.
type ModulePermissions struct {
	// Module name
	name string
	// All registered module permission
	perms Permissions
	// Other module - permission matching map
	modsPerms map[string]map[Permission]bool
}

func (m *ModulePermissions) String() string {
	strBuilder := strings.Builder{}

	strBuilder.WriteString(fmt.Sprintf("Module %q permissions:\n", m.name))
	strBuilder.WriteString(fmt.Sprintf("  Availdable perms: [%s]", m.perms.String()))
	for module, perms := range m.modsPerms {
		strBuilder.WriteString(fmt.Sprintf("\n  - %s:\n", module))
		permIdx := 0
		for perm := range perms {
			strBuilder.WriteString(fmt.Sprintf("    > %s", perm))
			if permIdx < len(perms)-1 {
				strBuilder.WriteString("\n")
			}
			permIdx++
		}
	}

	return strBuilder.String()
}

// AddModulePermission checks that {perms} are supported by target module and registers them.
func (m *ModulePermissions) AddModulePermission(requester RequestModulePermissions) error {
	moduleName, modulePerms := requester()

	if moduleName == "" {
		return fmt.Errorf("module %q: requester moduleName: empty", m.name)
	}
	if len(modulePerms) == 0 {
		return fmt.Errorf("module %q: requester modulePerms: empty", m.name)
	}

	modPermsMap := make(map[Permission]bool, len(modulePerms))
	for _, perm := range modulePerms {
		found := false
		for _, allowedPerm := range m.perms {
			if perm == allowedPerm {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("permission %q for module %q: not allowed by target module %q", moduleName, perm.String(), m.name)
		}
		modPermsMap[perm] = true
	}
	m.modsPerms[moduleName] = modPermsMap

	return nil
}

// AutoAddRequester wraps AddModulePermission and panics on failure.
func (m *ModulePermissions) AutoAddRequester(requester RequestModulePermissions) {
	if err := m.AddModulePermission(requester); err != nil {
		panic(err)
	}
}

// Check that {expectedPerm} for {moduleName} is allowed by the target module.
func (m *ModulePermissions) Check(moduleName string, expectedPerm Permission) error {
	modPermsMap, ok := m.modsPerms[moduleName]
	if !ok {
		return fmt.Errorf("target module %q: %q module is not supported", m.name, moduleName)
	}

	if !modPermsMap[expectedPerm] {
		return fmt.Errorf("target module %q: permission %q is not allowed for %q module", m.name, expectedPerm.String(), moduleName)
	}

	return nil
}

// AutoCheck wraps Check with caller module getter and panics is perm is not allowed.
func (m *ModulePermissions) AutoCheck(expectedPerm Permission) {
	requester, caller := helpers.Caller(1)

	if caller.Module == requester.Module {
		return
	}

	if err := m.Check(caller.Module, expectedPerm); err != nil {
		panic(err)
	}
}

// NewModulePermissions creates a new ModulePermission for target module.
func NewModulePermissions(targetModuleName string, targetPerms Permissions) ModulePermissions {
	return ModulePermissions{
		name:      targetModuleName,
		perms:     targetPerms,
		modsPerms: make(map[string]map[Permission]bool),
	}
}
