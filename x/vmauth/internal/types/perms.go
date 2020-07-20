package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/ccstorage"
)

// RequestCCStoragePerms returns module perms used by this module.
func RequestCCStoragePerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = "vmauth"
		modulePerms = perms.Permissions{
			ccstorage.PermCCUpdater,
			ccstorage.PermCCReader,
			ccstorage.PermCCResUpdater,
		}
		return
	}
}
