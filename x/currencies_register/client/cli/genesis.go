package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

const (
	flagClientHome = "home-client"
)

// AddGenesisAccountCmd allowing to add currency info into genesis with node.
func AddGenesisCurrencyInfo(ctx *server.Context, cdc *codec.Codec,
	defaultNodeHome, defaultClientHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-currency-info [denom] [decimals] [totalSupply] [path]",
		Short: "Add currency info to genesis.json",
		Args:  cobra.ExactArgs(4),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			denom := args[0]
			if err := sdk.ValidateDenom(denom); err != nil {
				return fmt.Errorf("incorrect denom %q: %v", denom, err)
			}

			decimals, err := strconv.ParseUint(args[1], 10, 8)
			if err != nil {
				return fmt.Errorf("can't parse decimals %q: %v", args[1], err)
			}

			totalSupply, isOk := sdk.NewIntFromString(args[2])
			if !isOk {
				return fmt.Errorf("can't parse total supply %q", totalSupply)
			}

			path, err := hex.DecodeString(args[3])
			if err != nil {
				return fmt.Errorf("path is not a correct hex %q: %v", args[3], err)
			}

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			genesisState := types.GenesisState{}
			cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisState)
			// find dublicated
			found := -1
			for i, genCurr := range genesisState.Currencies {
				if genCurr.Denom == denom {
					found = i
					break
				}
			}

			if found >= 0 {
				genesisState.Currencies[found].Path = hex.EncodeToString(path)
				genesisState.Currencies[found].TotalSupply = totalSupply
				genesisState.Currencies[found].Decimals = uint8(decimals)
			} else {
				genesisState.Currencies = append(genesisState.Currencies, types.GenesisCurrency{
					Path:        hex.EncodeToString(path),
					Denom:       denom,
					Decimals:    uint8(decimals),
					TotalSupply: totalSupply,
				})
			}

			genesisStateBz := cdc.MustMarshalJSON(genesisState)
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
