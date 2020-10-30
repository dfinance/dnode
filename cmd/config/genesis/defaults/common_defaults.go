// Commonly used default values.
// Moved to a separate pkg to prevent circular dependency.
package defaults

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MainDenom              = "xfi"
	StakingDenom           = "xfi"
	LiquidityProviderDenom = "lpt"

	// Min TX fee
	FeeAmount = "100000000000000" // 0.0001
	// Governance: deposit amount
	GovMinDepositAmount = "1000000000000000000000" // 1000.0
	// Staking: min self-delegation amount
	MinSelfDelegationAmount = "2500000000000000000000" // 2500.0
	// Staking: max self-delegation amount
	MaxSelfDelegationAmount = "10000000000000000000000" // 10000.0
	// Crisis: invariants check TX fee
	InvariantCheckAmount = "1000000000000000000000" // 1000.0
	// Distribution: PublicTreasuryPool capacity
	PublicTreasuryPoolAmount = "250000000000000000000000" // 250000.0

	MaxGas = 10000000
)

var (
	FeeCoin                    sdk.Coin
	GovMinDepositCoin          sdk.Coin
	MinSelfDelegationCoin      sdk.Coin
	MaxSelfDelegationCoin      sdk.Coin
	InvariantCheckCoin         sdk.Coin
	PublicTreasuryPoolCapacity sdk.Int
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

	if value, ok := sdk.NewIntFromString(MaxSelfDelegationAmount); !ok {
		panic("staking defaults: MaxSelfDelegationCoin conversion failed")
	} else {
		MaxSelfDelegationCoin = sdk.NewCoin(StakingDenom, value)
	}

	if value, ok := sdk.NewIntFromString(InvariantCheckAmount); !ok {
		panic("crisis defaults: InvariantCheckAmount conversion failed")
	} else {
		InvariantCheckCoin = sdk.NewCoin(MainDenom, value)
	}

	if value, ok := sdk.NewIntFromString(PublicTreasuryPoolAmount); !ok {
		panic("distribution defaults: PublicTreasuryPoolAmount conversion failed")
	} else {
		PublicTreasuryPoolCapacity = value
	}
}
