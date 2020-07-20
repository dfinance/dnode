package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	vmExport "github.com/dfinance/dnode/x/vm/export"
)

const (
	// Init genesis
	PermInit perms.Permission = ModuleName + "PermInit"
	// Create a new currency
	PermCCCreator perms.Permission = ModuleName + "PermCCCreator"
	// Update currency supply
	PermCCUpdater perms.Permission = ModuleName + "PermCCUpdater"
	// Read currency and resources
	PermCCReader perms.Permission = ModuleName + "PermCCReader"
	// Update currency VM resources
	PermCCResUpdater perms.Permission = ModuleName + "PermCCResUpdater"
)

var (
	AvailablePermissions = perms.Permissions{PermInit, PermCCCreator, PermCCUpdater, PermCCReader, PermCCResUpdater}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}

// RequestVMStoragePerms returns module perms used by this module.
func RequestVMStoragePerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			vmExport.PermStorageReader,
			vmExport.PermStorageWriter,
		}
		return
	}
}
