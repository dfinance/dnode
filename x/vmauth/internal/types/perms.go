package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	ccsClient "github.com/dfinance/dnode/x/ccstorage/client"
)

// RequestCCStoragePerms returns module perms used by this module.
func RequestCCStoragePerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = "vmauth"
		modulePerms = perms.Permissions{
			ccsClient.PermUpdate,
			ccsClient.PermRead,
			ccsClient.PermResUpdate,
		}
		return
	}
}
