package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmTypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

const (
	flagOverwrite = "overwrite"
	maxGas        = 10000000
)

type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string,
	appMessage json.RawMessage) printInfo {

	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

func displayInfo(cdc *codec.Codec, info printInfo) error {
	out, err := codec.MarshalJSONIndent(cdc, info)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", string(sdk.MustSortJSON(out)))
	return err
}

// InitCmd returns a command that initializes all files needed for Tendermint
// and the respective application.
func InitCmd(ctx *server.Context, cdc *codec.Codec, mbm module.BasicManager,
	defaultNodeHome string) *cobra.Command { // nolint: golint
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			chainID := viper.GetString(flags.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", tmrand.Str(6))
			}

			nodeID, _, err := genutil.InitializeNodeValidatorFiles(config)
			if err != nil {
				return err
			}

			config.Moniker = args[0]

			// Prepare genesis state
			appGenState, err := OverrideGenesisStateDefaults(cdc, mbm.DefaultGenesis())
			if err != nil {
				return fmt.Errorf("app genesis state overwrite: %v", err)
			}

			appState, err := codec.MarshalJSONIndent(cdc, appGenState)
			if err != nil {
				return errors.Wrap(err, "failed to marshall app genesis state")
			}

			// Prepare genesis file
			genFile := config.GenesisFile()
			if !viper.GetBool(flagOverwrite) && tmos.FileExists(genFile) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			genDoc := &tmTypes.GenesisDoc{}
			if _, err := os.Stat(genFile); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				genDoc, err = tmTypes.GenesisDocFromFile(genFile)
				if err != nil {
					return errors.Wrap(err, "failed to read genesis doc from file")
				}
			}

			genDoc.ChainID = chainID
			genDoc.Validators = nil
			genDoc.AppState = appState

			// Setup max gas
			if genDoc.ConsensusParams == nil {
				genDoc.ConsensusParams = tmTypes.DefaultConsensusParams()
			}
			genDoc.ConsensusParams.Block.MaxGas = maxGas

			if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
				return errors.Wrap(err, "Failed to export gensis file")
			}

			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "", appState)

			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
			return displayInfo(cdc, toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")

	return cmd
}

// OverrideGenesisStateDefaults takes default app genesis state and overwrites Cosmos SDK / Dfinance params.
func OverrideGenesisStateDefaults(cdc *codec.Codec, genState map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	// Staking module params
	{
		moduleName, moduleState := staking.ModuleName, staking.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		minSelfDelegation, ok := sdk.NewIntFromString(DefMinSelfDelegation)
		if !ok {
			return nil, fmt.Errorf("%s module: default minSelfDelegation convertion failed: %s", moduleName, DefMinSelfDelegation)
		}

		moduleState.Params.BondDenom = StakingDenom
		moduleState.Params.LPDenom = LiquidityProviderDenom
		moduleState.Params.MinSelfDelegationLvl = minSelfDelegation

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	// Mint module params
	{
		moduleName, moduleState := mint.ModuleName, mint.GenesisState{}
		if err := cdc.UnmarshalJSON(genState[moduleName], &moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON unmarshal: %v", moduleName, err)
		}

		moduleState.Params.MintDenom = StakingDenom
		//
		moduleState.Params.InflationMax = sdk.NewDecWithPrec(50, 2)   // 50%
		moduleState.Params.InflationMin = sdk.NewDecWithPrec(1776, 4) // 17.76%
		//
		moduleState.Params.FeeBurningRatio = sdk.NewDecWithPrec(50, 2)           // 50%
		moduleState.Params.InfPwrBondedLockedRatio = sdk.NewDecWithPrec(4, 1)    // 40%
		moduleState.Params.FoundationAllocationRatio = sdk.NewDecWithPrec(45, 2) // 45%
		//
		moduleState.Params.AvgBlockTimeWindow = 100 // 100 blocks

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", err)
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
		moduleState.Params.PublicTreasuryPoolCapacity = sdk.NewInt(250000) // 250K (doesn't include currency decimals)
		//
		moduleState.Params.BaseProposerReward = sdk.NewDecWithPrec(1, 2)  // 1%
		moduleState.Params.BonusProposerReward = sdk.NewDecWithPrec(4, 2) // 4%
		//
		moduleState.Params.WithdrawAddrEnabled = true

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", err)
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

		moduleState.DepositParams.MinDeposit = sdk.NewCoins(GovMinDeposit)

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", err)
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

		invariantCheckFeeAmount, ok := sdk.NewIntFromString(DefaultFeeAmount)
		if !ok {
			return nil, fmt.Errorf("%s module: invariant check fee convertion failed: %s", moduleName, DefaultFeeAmount)
		}

		moduleState.ConstantFee.Denom = MainDenom
		moduleState.ConstantFee.Amount = invariantCheckFeeAmount

		if moduleStateBz, err := cdc.MarshalJSON(moduleState); err != nil {
			return nil, fmt.Errorf("%s module: JSON marshal: %v", err)
		} else {
			genState[moduleName] = moduleStateBz
		}
	}

	return genState, nil
}
