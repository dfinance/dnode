package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

const (
	flagClientHome = "home-client"
)

// AddOracleNomeneesCmd returns add-oracle-nomenees cobra Command.
func AddOracleNomeneesCmd(ctx *server.Context, cdc *codec.Codec,
	defaultNodeHome, defaultClientHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-oracle-nomenees [address1,address2...]",
		Short: "Add oracle nomenees to genesis.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			addresses := strings.Split(args[0], ",")
			for i, a := range addresses {
				if _, err := sdk.AccAddressFromBech32(a); err != nil {
					return fmt.Errorf("%q address at index %d: %w", a, i, err)
				}
			}

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			// add genesis account to the app state
			var genesisOracle types.GenesisState

			cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisOracle)
			var addedNomenees []string
			for _, n := range addresses {
				var found bool
				for _, nn := range genesisOracle.Params.Nominees {
					if n == nn {
						addedNomenees = append(addedNomenees, nn)
						found = true
					}
				}
				if !found {
					addedNomenees = append(addedNomenees, n)
				}
			}

			genesisOracle.Params.Nominees = addedNomenees

			genesisStateBz := cdc.MustMarshalJSON(genesisOracle)
			appState[types.ModuleName] = genesisStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}

			// export app state
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")
	return cmd
}
