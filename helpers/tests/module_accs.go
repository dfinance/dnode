package tests

import (
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

var (
	// Module account permissions
	MAccPerms map[string][]string = map[string][]string{
		auth.FeeCollectorName: nil,
		"orders":              {supply.Burner},
	}
)

func ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range MAccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}
