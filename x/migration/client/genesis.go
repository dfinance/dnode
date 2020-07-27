package client

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	tmTypes "github.com/tendermint/tendermint/types"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/migration/internal/types"
)

const (
	flagGenesisTime = "genesis-time"
	flagChainID     = "chain-id"
)

// MigrateGenesisCmd returns a command to execute genesis state migration.
func MigrateGenesisCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate [targetVersion] [genesisFile]",
		Short:   "Migrate genesis state to a specified target version",
		Example: "migrate v0.7 ./genesis.json --chain-id=testnet --genesis-time=2019-04-22T17:00:00Z",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// parse inputs
			targetVersion := args[0]
			migrationHandler := types.MigrationMap[targetVersion]
			if migrationHandler == nil {
				return helpers.BuildError("targetVersion", targetVersion, helpers.ParamTypeCliArg, "migration handler not found")
			}

			genFile := args[1]
			if err := helpers.CheckFileExists("genesisFile", genFile, helpers.ParamTypeCliArg); err != nil {
				return err
			}

			updChainID := cmd.Flag(flagChainID).Value.String()

			var updGenesisTime time.Time
			if value := cmd.Flag(flagGenesisTime).Value.String(); value != "" {
				if err := updGenesisTime.UnmarshalText([]byte(value)); err != nil {
					return helpers.BuildError(flagGenesisTime, value, helpers.ParamTypeCliFlag, fmt.Sprintf("failed to unmarshal time: %v", err))
				}
			}

			// retrieve the app state
			genDoc, err := tmTypes.GenesisDocFromFile(genFile)
			if err != nil {
				return fmt.Errorf("reading initial appState (path %q): %w", genFile, err)
			}

			var appStateInitial genutil.AppMap
			if err := cdc.UnmarshalJSON(genDoc.AppState, &appStateInitial); err != nil {
				return fmt.Errorf("initial appState JSON unmarshal: %w", err)
			}

			// migrate
			appStateMigrated, err := migrationHandler(appStateInitial)
			if err != nil {
				return fmt.Errorf("migration to %q failed: %w", targetVersion, err)
			}

			// update the genesisDoc and print the result
			appStateMigratedBz, err := cdc.MarshalJSON(appStateMigrated)
			if err != nil {
				return fmt.Errorf("migrated appState JSON marshal: %w", err)
			}
			genDoc.AppState = appStateMigratedBz

			if updChainID != "" {
				genDoc.ChainID = updChainID
			}
			if !updGenesisTime.IsZero() {
				genDoc.GenesisTime = updGenesisTime
			}

			bz, err := cdc.MarshalJSONIndent(genDoc, "", "  ")
			if err != nil {
				return fmt.Errorf("genesisDoc JSON marshal: %w", err)
			}

			sortedBz, err := sdk.SortJSON(bz)
			if err != nil {
				return fmt.Errorf("genesisDoc JSON sorting: %w", err)
			}

			fmt.Println(string(sortedBz))

			return nil
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"target migration version (Golang semver format without build version)",
		"path to an exported genesis state file for current state version",
	})
	cmd.Flags().String(flagGenesisTime, "", "override genesis_time")
	cmd.Flags().String(flagChainID, "", "override chain_id")

	return cmd
}
