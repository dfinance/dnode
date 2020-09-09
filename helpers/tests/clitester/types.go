package clitester

import (
	"encoding/json"
	"os"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/currencies"
	"github.com/dfinance/dnode/x/genaccounts"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/multisig"
	"github.com/dfinance/dnode/x/oracle"
	"github.com/dfinance/dnode/x/orderbook"
	"github.com/dfinance/dnode/x/orders"
	"github.com/dfinance/dnode/x/poa"
	"github.com/dfinance/dnode/x/vm"
)

const (
	DenomXFI  = "xfi"
	DenomSXFI = "sxfi"
	DenomETH  = "eth"
	DenomBTC  = "btc"
	DenomUSDT = "usdt"
	//
	DefaultGas = 500000
)

type StringPair struct {
	Key   string
	Value string
}

var DefVmWriteSetsPath = "${GOPATH}/src/github.com/dfinance/dnode/x/vm/internal/keeper/genesis_ws.json"

func init() {
	vmWriteSetPath := os.Getenv("VMWSPATH")
	if vmWriteSetPath != "" {
		DefVmWriteSetsPath = vmWriteSetPath
	}
}

var EthAddresses = []string{
	"0x82A978B3f5962A5b0957d9ee9eEf472EE55B42F1",
	"0x7d577a597B2742b498Cb5Cf0C26cDCD726d39E6e",
	"0xDCEceAF3fc5C0a63d195d69b1A90011B7B19650D",
	"0x598443F1880Ef585B21f1d7585Bd0577402861E5",
	"0x13cBB8D99C6C4e0f2728C7d72606e78A29C4E224",
	"0x77dB2BEBBA79Db42a978F896968f4afCE746ea1F",
	"0x24143873e0E0815fdCBcfFDbe09C979CbF9Ad013",
	"0x10A1c1CB95c92EC31D3f22C66Eef1d9f3F258c6B",
	"0xe0FC04FA2d34a66B779fd5CEe748268032a146c0",
}

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
	ccstorage.AppModuleBasic{},
	currencies.AppModuleBasic{},
	multisig.AppModuleBasic{},
	oracle.AppModuleBasic{},
	vm.AppModuleBasic{},
	orders.AppModuleBasic{},
	markets.AppModuleBasic{},
	orderbook.AppModuleBasic{},
	gov.AppModuleBasic{},
)
