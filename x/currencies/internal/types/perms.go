package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	ccsClient "github.com/dfinance/dnode/x/ccstorage/client"
)

const (
	// Init genesis
	PermInit perms.Permission = ModuleName + "PermInit"
	// Create a new currency
	PermCreate perms.Permission = ModuleName + "PermCreate"
	// Issue currency amount
	PermIssue perms.Permission = ModuleName + "PermIssue"
	// Withdraw currency amount
	PermWithdraw perms.Permission = ModuleName + "PermWithdraw"
	// Read Issue / Withdraw
	PermRead perms.Permission = ModuleName + "PermRead"
)

var (
	AvailablePermissions = perms.Permissions{PermInit, PermCreate, PermIssue, PermWithdraw, PermRead}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}

// RequestCCStoragePerms returns module perms used by this module.
func RequestCCStoragePerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			ccsClient.PermCreate,
			ccsClient.PermUpdate,
			ccsClient.PermRead,
		}
		return
	}
}
