package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/markets"
)

const (
	// Post a new order
	PermOrderPost perms.Permission = ModuleName + "PermOrderPost"
	// Revoke order
	PermOrderRevoke perms.Permission = ModuleName + "PermOrderRevoke"
	// Read order / orders
	PermReader perms.Permission = ModuleName + "PermReader"
	// Lock order coins
	PermOrderLock perms.Permission = ModuleName + "PermOrderLock"
	// Unlock order coins
	PermOrderUnlock perms.Permission = ModuleName + "PermOrderUnlock"
	// Execute order fills
	PermExecFills perms.Permission = ModuleName + "PermExecFills"
)

var (
	AvailablePermissions = perms.Permissions{PermOrderPost, PermOrderRevoke, PermReader, PermOrderLock, PermOrderUnlock, PermExecFills}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}

// RequestMarketsPerms returns module perms used by this module.
func RequestMarketsPerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			markets.PermReader,
		}
		return
	}
}
