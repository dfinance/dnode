package genesis

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/dfinance/dnode/cmd/config/genesis/defaults"
	"github.com/dfinance/dnode/x/multisig"
	"github.com/dfinance/dnode/x/oracle"
	"github.com/dfinance/dnode/x/poa"
)

const (
	AvgYearDur = time.Duration(60*60*8766) * time.Second // 365.25 days
	DayDur     = 24 * time.Hour
)

// OverrideGenesisStateDefaults takes default app genesis state and overwrites Cosmos SDK / Dfinance params.
func OverrideGenesisStateDefaults(cdc *codec.Codec, genState map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	// Mint module params
	{
		moduleName, moduleState := mint.ModuleName, mint.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.Params.MintDenom = defaults.StakingDenom
		//
		moduleState.Params.InflationMax = sdk.NewDecWithPrec(50, 2)   // 50%
		moduleState.Params.InflationMin = sdk.NewDecWithPrec(1776, 4) // 17.76%
		//
		moduleState.Params.FeeBurningRatio = sdk.NewDecWithPrec(50, 2)           // 50%
		moduleState.Params.InfPwrBondedLockedRatio = sdk.NewDecWithPrec(4, 1)    // 40%
		moduleState.Params.FoundationAllocationRatio = sdk.NewDecWithPrec(15, 2) // 15%
		//
		moduleState.Params.AvgBlockTimeWindow = 100 // 100 blocks

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", moduleName, err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	// Distribution module params
	{
		moduleName, moduleState := distribution.ModuleName, distribution.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.Params.ValidatorsPoolTax = sdk.NewDecWithPrec(4825, 4)         // 48.25%
		moduleState.Params.LiquidityProvidersPoolTax = sdk.NewDecWithPrec(4825, 4) // 48.25%
		moduleState.Params.PublicTreasuryPoolTax = sdk.NewDecWithPrec(15, 3)       // 1.5%
		moduleState.Params.HARPTax = sdk.NewDecWithPrec(2, 2)                      // 2%
		//
		moduleState.Params.PublicTreasuryPoolCapacity = defaults.PublicTreasuryPoolCapacity // 250000.0
		//
		moduleState.Params.BaseProposerReward = sdk.NewDecWithPrec(1, 2)  // 1%
		moduleState.Params.BonusProposerReward = sdk.NewDecWithPrec(4, 2) // 4%
		//
		moduleState.Params.LockedRatio = sdk.NewDecWithPrec(5, 1) // 50%
		moduleState.Params.LockedDuration = AvgYearDur            // 1 year
		//
		moduleState.Params.WithdrawAddrEnabled = true
		moduleState.Params.FoundationNominees = []sdk.AccAddress{}

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", moduleName, err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	// Staking module params
	{
		moduleName, moduleState := staking.ModuleName, staking.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.Params.UnbondingTime = 7 * DayDur
		moduleState.Params.ScheduledUnbondDelayTime = 3 * DayDur
		//
		moduleState.Params.MaxValidators = 31
		moduleState.Params.MaxEntries = 7
		moduleState.Params.HistoricalEntries = 0
		//
		moduleState.Params.BondDenom = defaults.StakingDenom
		moduleState.Params.LPDenom = defaults.LiquidityProviderDenom
		moduleState.Params.LPDistrRatio = sdk.NewDecWithPrec(1, 0) // 100%
		//
		moduleState.Params.MinSelfDelegationLvl = defaults.MinSelfDelegationCoin.Amount // 2500.0
		moduleState.Params.MaxDelegationsRatio = sdk.NewDecWithPrec(10, 0)              // 10.0

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", moduleName, err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	// Slashing module params
	{
		moduleName, moduleState := slashing.ModuleName, slashing.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.Params.SignedBlocksWindow = 100
		moduleState.Params.MinSignedPerWindow = sdk.NewDecWithPrec(5, 1)
		//
		moduleState.Params.DowntimeJailDuration = 10 * time.Minute
		//
		moduleState.Params.SlashFractionDoubleSign = sdk.NewDec(1).Quo(sdk.NewDec(20)) // 1.0 / 20.0
		moduleState.Params.SlashFractionDowntime = sdk.NewDec(1).Quo(sdk.NewDec(100))  // 1.0 / 100.0

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", moduleName, err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	// Evidence module params
	{
		moduleName, moduleState := evidence.ModuleName, evidence.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.Params.MaxEvidenceAge = 2 * time.Minute

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", moduleName, err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	// Crisis module params
	{
		moduleName, moduleState := crisis.ModuleName, crisis.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.ConstantFee.Denom = defaults.InvariantCheckCoin.Denom   // xfi
		moduleState.ConstantFee.Amount = defaults.InvariantCheckCoin.Amount // 1000.0

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", moduleName, err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	// Gov module params
	{
		moduleName, moduleState := gov.ModuleName, gov.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.VotingParams.VotingPeriod = 6 * time.Hour
		//
		moduleState.TallyParams.Quorum = sdk.NewDecWithPrec(334, 3)  // 33.4%
		moduleState.TallyParams.Threshold = sdk.NewDecWithPrec(5, 1) // 50%
		moduleState.TallyParams.Veto = sdk.NewDecWithPrec(334, 3)    // 33.4%
		//
		moduleState.DepositParams.MinDeposit = sdk.NewCoins(defaults.GovMinDepositCoin) // 1000.0sxfi
		moduleState.DepositParams.MaxDepositPeriod = 3 * time.Hour

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", moduleName, err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	// Auth module params
	{
		moduleName, moduleState := auth.ModuleName, auth.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.Params.MaxMemoCharacters = 256
		moduleState.Params.TxSigLimit = 7
		moduleState.Params.TxSizeCostPerByte = 10
		moduleState.Params.SigVerifyCostED25519 = 590
		moduleState.Params.SigVerifyCostSecp256k1 = 1000

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", moduleName, err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	// Bank module genesis
	{
		moduleName, moduleState := bank.ModuleName, bank.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.SendEnabled = true

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", moduleName, err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	// PoA module params
	{
		moduleName, moduleState := poa.ModuleName, poa.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.Parameters.MaxValidators = 11
		moduleState.Parameters.MinValidators = 1

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", moduleName, err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	// MultiSig module params
	{
		moduleName, moduleState := multisig.ModuleName, multisig.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.Parameters.IntervalToExecute = 86400 // 6 days approx.

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", moduleName, err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	// Oracle module params
	{
		moduleName, moduleState := oracle.ModuleName, oracle.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.Params.PostPrice.ReceivedAtDiffInS = 360 // 1 hour

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", moduleName, err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	return genState, nil
}
