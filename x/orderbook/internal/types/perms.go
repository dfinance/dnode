package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/orders"
)

const (
	// Read history item / items
	PermHistoryReader perms.Permission = ModuleName + "PermHistoryReader"
	// Write history item
	PermHistoryWriter perms.Permission = ModuleName + "PermHistoryWriter"
	// Read orders
	PermOrdersRead perms.Permission = ModuleName + "PermOrdersRead"
	// Execute order fills
	PermExecFills perms.Permission = ModuleName + "PermExecFills"
)

var (
	AvailablePermissions = perms.Permissions{PermHistoryReader, PermHistoryWriter, PermOrdersRead, PermExecFills}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}

// RequestOrdersPerms returns module perms used by this module.
func RequestOrdersPerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			orders.PermReader,
			orders.PermExecFills,
		}
		return
	}
}
