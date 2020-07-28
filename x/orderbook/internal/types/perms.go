package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	ordersClient "github.com/dfinance/dnode/x/orders/client"
)

const (
	// Init genesis
	PermInit perms.Permission = ModuleName + "PermInit"
	// Export genesis
	PermExport perms.Permission = ModuleName + "PermExport"
	// Read history item / items
	PermHistoryRead perms.Permission = ModuleName + "PermHistoryRead"
	// Write history item
	PermHistoryWrite perms.Permission = ModuleName + "PermHistoryWrite"
	// Read orders
	PermOrdersRead perms.Permission = ModuleName + "PermOrdersRead"
	// Execute order fills
	PermExecFill perms.Permission = ModuleName + "PermExecFill"
)

var (
	AvailablePermissions = perms.Permissions{
		PermExport,
		PermInit,
		PermHistoryRead,
		PermHistoryWrite,
		PermOrdersRead,
		PermExecFill,
	}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}

// RequestOrdersPerms returns module perms used by this module.
func RequestOrdersPerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			ordersClient.PermRead,
			ordersClient.PermExecFill,
		}
		return
	}
}
