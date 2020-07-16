package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/markets"
)

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
