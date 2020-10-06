// Commonly used default values.
// Moved to a separate pkg to prevent circular dependency.
package defaults

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MainDenom              = "xfi"
	StakingDenom           = "sxfi"
	LiquidityProviderDenom = "lpt"

	FeeAmount               = "100000000000000"        // 0.0001
	GovMinDepositAmount     = "1000000000000000000000" // 1000.0
	MinSelfDelegationAmount = "2500000000000000000000" // 2500.0
	InvariantCheckAmount    = "1000000000000000000000" // 1000.0

	MaxGas = 10000000
)

var (
	FeeCoin               sdk.Coin
	GovMinDepositCoin     sdk.Coin
	MinSelfDelegationCoin sdk.Coin
	InvariantCheckCoin    sdk.Coin
)

func init() {
	if value, ok := sdk.NewIntFromString(FeeAmount); !ok {
		panic("defaults: FeeAmount conversion failed")
	} else {
		FeeCoin = sdk.NewCoin(MainDenom, value)
	}

	if value, ok := sdk.NewIntFromString(GovMinDepositAmount); !ok {
		panic("governance defaults: GovMinDepositAmount conversion failed")
	} else {
		GovMinDepositCoin = sdk.NewCoin(StakingDenom, value)
	}

	if value, ok := sdk.NewIntFromString(MinSelfDelegationAmount); !ok {
		panic("staking defaults: MinSelfDelegationAmount conversion failed")
	} else {
		MinSelfDelegationCoin = sdk.NewCoin(StakingDenom, value)
	}

	if value, ok := sdk.NewIntFromString(InvariantCheckAmount); !ok {
		panic("crisis defaults: InvariantCheckAmount conversion failed")
	} else {
		InvariantCheckCoin = sdk.NewCoin(MainDenom, value)
	}
}
