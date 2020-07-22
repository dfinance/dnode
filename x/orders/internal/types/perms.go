package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	marketsClient "github.com/dfinance/dnode/x/markets/client"
)

const (
	// Post a new order
	PermOrderPost perms.Permission = ModuleName + "PermOrderPost"
	// Revoke order
	PermOrderRevoke perms.Permission = ModuleName + "PermOrderRevoke"
	// Read order / orders
	PermRead perms.Permission = ModuleName + "PermRead"
	// Lock order coins
	PermOrderLock perms.Permission = ModuleName + "PermOrderLock"
	// Unlock order coins
	PermOrderUnlock perms.Permission = ModuleName + "PermOrderUnlock"
	// Execute order fills
	PermExecFill perms.Permission = ModuleName + "PermExecFill"
)

var (
	AvailablePermissions = perms.Permissions{PermOrderPost, PermOrderRevoke, PermRead, PermOrderLock, PermOrderUnlock, PermExecFill}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}

// RequestMarketsPerms returns module perms used by this module.
func RequestMarketsPerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			marketsClient.PermRead,
		}
		return
	}
}
