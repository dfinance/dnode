package clitester

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/dfinance/dnode/x/currencies"
	"github.com/dfinance/dnode/x/multisig"
	"github.com/dfinance/dnode/x/oracle"
	"github.com/dfinance/dnode/x/poa"
	"github.com/dfinance/dnode/x/vm"
)

type GenesisState map[string]json.RawMessage

var ModuleBasics = module.NewBasicManager(
	genaccounts.AppModuleBasic{},
	genutil.AppModuleBasic{},
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	staking.AppModuleBasic{},
	distribution.AppModuleBasic{},
	params.AppModuleBasic{},
	slashing.AppModuleBasic{},
	supply.AppModuleBasic{},
	poa.AppModuleBasic{},
	currencies.AppModuleBasic{},
	multisig.AppModuleBasic{},
	oracle.AppModuleBasic{},
	vm.AppModuleBasic{},
)
