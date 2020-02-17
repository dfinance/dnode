package cli

import (
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"wings-blockchain/x/vm/internal/types"
)

// Get query commands for VM module.
func GetQueriesCmd(cdc *codec.Codec) *cobra.Command {
	queries := &cobra.Command{
		Use:   "vm",
		Short: "VM query commands, include compiler",
	}

	queries.AddCommand(
		client.GetCommands(
			GetData("vm", cdc),
			CompileScript("vm", cdc),
			CompileModule("vm", cdc),
		)...,
	)

	return queries
}

// Get data from data source by access path.
func GetData(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "get_data [address] [path]",
		Short:   "get_data from data source storage by address and path",
		Example: "get data 0000000000000000000000000000000000000000000000000000000000000000 0019b01c2cf3c2160a43e4dcad70e3e5d18151cc38de7a1d1067c6031bfa0ae4d9",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// extract data
			address, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			path, err := hex.DecodeString(args[1])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.QueryAccessPath{
				Address: address,
				Path:    path,
			})
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(
				fmt.Sprintf("custom/%s/value", queryRoute),
				bz)
			if err != nil {
				return err
			}

			out := types.QueryValueResp{Value: hex.EncodeToString(res)}
			return cliCtx.PrintOutput(out)
		},
	}
}

// Compile Mvir script.
func CompileScript(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "compile_script [mvirFile]",
		Short: "compile script using source code from mvir file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

// Compile Mvir module.
func CompileModule(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "compile_module [account] [mvirFile]",
		Short: "compile module connected to account, using source code from mvir file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
