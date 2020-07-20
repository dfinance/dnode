package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	poaExport "github.com/dfinance/dnode/x/poa/export"
)

const (
	// Init genesis
	PermInit perms.Permission = ModuleName + "PermInit"
	// Read POA data
	PermPoaReader perms.Permission = ModuleName + "PermPoaReader"
	// Read calls, handlers
	PermReader perms.Permission = ModuleName + "PermReader"
	// Add/update calls
	PermWriter perms.Permission = ModuleName + "PermWriter"
)

var (
	AvailablePermissions = perms.Permissions{PermInit, PermPoaReader, PermReader, PermWriter}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}

// RequestPoaPerms returns module perms used by this module.
func RequestPoaPerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			poaExport.PermReader,
		}
		return
	}
}
