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

			genFile := config.GenesisFile()
			if !viper.GetBool(flagOverwrite) && tmos.FileExists(genFile) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			appGenState := mbm.DefaultGenesis()

			// Change default staking denom, minSelfDelegation
			minSelfDelegation, ok := sdk.NewIntFromString(DefMinSelfDelegation)
			if !ok {
				return fmt.Errorf("staking genState: default minSelfDelegation convertion failed: %s", DefMinSelfDelegation)
			}

			stakingDataBz := appGenState[staking.ModuleName]
			var stakingGenState staking.GenesisState

			cdc.MustUnmarshalJSON(stakingDataBz, &stakingGenState)
			stakingGenState.Params.BondDenom = SXFIDenom
			stakingGenState.Params.MinSelfDelegationLvl = minSelfDelegation
			appGenState[staking.ModuleName] = cdc.MustMarshalJSON(stakingGenState)

			// Change default mint params
			mintDataBz := appGenState[mint.ModuleName]
			var mintGenState mint.GenesisState

			cdc.MustUnmarshalJSON(mintDataBz, &mintGenState)
			mintGenState.Params.MintDenom = MainDenom
			//
			mintGenState.Params.InflationMax = sdk.NewDecWithPrec(50, 2)   // 50%
			mintGenState.Params.InflationMin = sdk.NewDecWithPrec(1776, 4) // 17.76%
			//
			mintGenState.Params.FeeBurningRatio = sdk.NewDecWithPrec(50, 2)           // 50%
			mintGenState.Params.InfPwrBondedLockedRatio = sdk.NewDecWithPrec(4, 1)    // 40%
			mintGenState.Params.FoundationAllocationRatio = sdk.NewDecWithPrec(45, 2) // 45%
			//
			mintGenState.Params.AvgBlockTimeWindow = 100 // 100 blocks
			appGenState[mint.ModuleName] = cdc.MustMarshalJSON(mintGenState)

			// Change default distribution params
			distDataBz := appGenState[distribution.ModuleName]
			var distGenState distribution.GenesisState

			cdc.MustUnmarshalJSON(distDataBz, &distGenState)
			distGenState.Params.ValidatorsPoolTax = sdk.NewDecWithPrec(4825, 4)         // 48.25%
			distGenState.Params.LiquidityProvidersPoolTax = sdk.NewDecWithPrec(4825, 4) // 48.25%
			distGenState.Params.PublicTreasuryPoolTax = sdk.NewDecWithPrec(15, 3)       // 1.5%
			distGenState.Params.HARPTax = sdk.NewDecWithPrec(2, 2)                      // 2%
			//
			distGenState.Params.PublicTreasuryPoolCapacity = sdk.NewInt(250000) // 250K (doesn't include currency decimals)
			//
			distGenState.Params.BaseProposerReward = sdk.NewDecWithPrec(1, 2)  // 1%
			distGenState.Params.BonusProposerReward = sdk.NewDecWithPrec(4, 2) // 4%
			//
			distGenState.Params.WithdrawAddrEnabled = true
			appGenState[distribution.ModuleName] = cdc.MustMarshalJSON(distGenState)

			// Change default minimal governance deposit coin
			govDataBz := appGenState[gov.ModuleName]
			var govGenState gov.GenesisState

			cdc.MustUnmarshalJSON(govDataBz, &govGenState)
			govGenState.DepositParams.MinDeposit = sdk.NewCoins(GovMinDeposit)
			appGenState[gov.ModuleName] = cdc.MustMarshalJSON(govGenState)

			// Change default crisis constant fee
			crisisDataBz := appGenState[crisis.ModuleName]
			var crisisGenState crisis.GenesisState

			cdc.MustUnmarshalJSON(crisisDataBz, &crisisGenState)
			defFeeAmount, _ := sdk.NewIntFromString(DefaultFeeAmount)
			crisisGenState.ConstantFee.Denom = MainDenom
			crisisGenState.ConstantFee.Amount = defFeeAmount
			appGenState[crisis.ModuleName] = cdc.MustMarshalJSON(crisisGenState)

			appState, err := codec.MarshalJSONIndent(cdc, appGenState)
			if err != nil {
				return errors.Wrap(err, "Failed to marshall default genesis state")
			}

			genDoc := &tmTypes.GenesisDoc{}
			if _, err := os.Stat(genFile); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				genDoc, err = tmTypes.GenesisDocFromFile(genFile)
				if err != nil {
					return errors.Wrap(err, "Failed to read genesis doc from file")
				}
			}

			genDoc.ChainID = chainID
			genDoc.Validators = nil
			genDoc.AppState = appState

			// Setup max gas.
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
