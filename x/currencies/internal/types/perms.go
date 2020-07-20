package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/ccstorage"
)

const (
	// Create a new currency
	PermCCCreator perms.Permission = ModuleName + "PermCCCreator"
	// Issue currency amount
	PermCCIssue perms.Permission = ModuleName + "PermCCIssue"
	// Withdraw currency amount
	PermCCWithdraw perms.Permission = ModuleName + "PermCCWithdraw"
	// Read Issue / Withdraw
	PermReader perms.Permission = ModuleName + "PermReader"
)

var (
	AvailablePermissions = perms.Permissions{PermCCCreator, PermCCIssue, PermCCWithdraw, PermReader}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}

// RequestCCStoragePerms returns module perms used by this module.
func RequestCCStoragePerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			ccstorage.PermCCCreator,
			ccstorage.PermCCUpdater,
			ccstorage.PermCCReader,
		}
		return
	}
}
