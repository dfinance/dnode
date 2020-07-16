package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/orders"
)

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
