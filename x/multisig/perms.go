package multisig

import (
	"github.com/dfinance/dnode/helpers/perms"
	poaClient "github.com/dfinance/dnode/x/poa/client"
)

// RequestPoaPerms returns module perms used by this module.
func RequestPoaPerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			poaClient.PermRead,
		}
		return
	}
}
