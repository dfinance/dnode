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

	dnTypes "github.com/dfinance/dnode/helpers/types"
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
		Short: "Add currency info to genesis state (non-token)",
		Args:  cobra.ExactArgs(4),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			// parse inputs
			denom, decimals, totalSupply, path, err := parseCurrencyArgs(args[0], args[1], args[2], args[3])
			if err != nil {
				return err
			}

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			genesisState := types.GenesisState{}
			cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisState)

			// find duplicated
			found := -1
			for i, genCurr := range genesisState.Currencies {
				if genCurr.Denom == denom {
					found = i
					break
				}
			}

			// update / add genesis state
			if found >= 0 {
				genesisState.Currencies[found].Path = hex.EncodeToString(path)
				genesisState.Currencies[found].TotalSupply = totalSupply
				genesisState.Currencies[found].Decimals = decimals
			} else {
				genesisState.Currencies = append(genesisState.Currencies, types.GenesisCurrency{
					Path:        hex.EncodeToString(path),
					Denom:       denom,
					Decimals:    decimals,
					TotalSupply: totalSupply,
				})
			}

			// update and export app state
			genesisStateBz := cdc.MustMarshalJSON(genesisState)
			appState[types.ModuleName] = genesisStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")

	return cmd
}

func parseCurrencyArgs(denomArg, decimalsArg, totalSupplyArg, pathArg string) (denom string, decimals uint8, totalSupply sdk.Int, path []byte, retErr error){
	if err := dnTypes.DenomFilter(denomArg); err != nil {
		retErr = fmt.Errorf("%s argument %q parse error: %w", "denom", denomArg, err)
		return
	} else {
		denom = denomArg
	}

	if v, err := strconv.ParseUint(decimalsArg, 10, 8); err != nil {
		retErr = fmt.Errorf("%s argument %q parse error: %w", "decimals", decimalsArg, err)
		return
	} else {
		decimals = uint8(v)
	}

	if v, ok := sdk.NewIntFromString(totalSupplyArg); !ok {
		retErr = fmt.Errorf("%s argument %q parse error: invalid big.Int", "totalSupply", totalSupplyArg)
		return
	} else {
		totalSupply = v
	}

	if v, err := hex.DecodeString(pathArg); err != nil {
		retErr = fmt.Errorf("%s argument %q parse error: %w", "path", pathArg, err)
		return
	} else {
		path = v
	}

	return
}
