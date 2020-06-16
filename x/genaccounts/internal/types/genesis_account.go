package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authExported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"
	supplyExported "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

type GenesisAccount struct {
	BaseAccount       auth.BaseAccount
	ModuleName        string
	ModulePermissions []string
}

// NewGenesisAccountRaw creates a new GenesisAccount object.
func NewGenesisAccountRaw(address sdk.AccAddress, coins sdk.Coins, moduleName string) GenesisAccount {
	acc := auth.NewBaseAccount(address, coins, nil, 0, 0)

	return GenesisAccount{
		BaseAccount: *acc,
		ModuleName:  moduleName,
	}
}

// NewGenesisAccountI creates a GenesisAccount instance from an Account interface.
func NewGenesisAccountI(acc authExported.Account) (GenesisAccount, error) {
	baseAcc := auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), nil, acc.GetAccountNumber(), acc.GetSequence())
	ga := GenesisAccount{BaseAccount: *baseAcc}

	if err := baseAcc.Validate(); err != nil {
		return ga, err
	}

	switch acc := acc.(type) {
	case supplyExported.ModuleAccountI:
		ga.ModuleName = acc.GetName()
		ga.ModulePermissions = acc.GetPermissions()
	}

	return ga, nil
}

func (ga GenesisAccount) ToAccount() authExported.Account {
	if ga.ModuleName != "" {
		return supply.NewModuleAccount(&ga.BaseAccount, ga.ModuleName, ga.ModulePermissions...)
	}

	return &ga.BaseAccount
}

func (ga GenesisAccount) Validate() error {
	return ga.BaseAccount.Validate()
}

type GenesisAccounts []GenesisAccount

// Check if state contains specified address.
func (ga GenesisAccounts) Contains(addr sdk.AccAddress) bool {
	for _, acc := range ga {
		if acc.BaseAccount.Address.Equals(addr) {
			return true
		}
	}

	return false
}