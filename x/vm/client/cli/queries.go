package cli

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/client/vm_client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// GetData returns query command that returns writeSet for VM accessPath.
func GetData(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get-data [address] [path]",
		Short:   "Get write set data from the storage by address and path",
		Example: "get-data wallet1jk4ld0uu6wdrj9t8u3gghm9jt583hxx7xp7he8 0019b01c2cf3c2160a43e4dcad70e3e5d18151cc38de7a1d1067c6031bfa0ae4d9",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			address, err := helpers.ParseSdkAddressParam("address", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			_, path, err := helpers.ParseHexStringParam("path", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare request
			bz, err := cdc.MarshalJSON(types.ValueReq{
				Address: common_vm.Bech32ToLibra(address),
				Path:    path,
			})
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValue), bz)
			if err != nil {
				return err
			}

			out := types.ValueResp{Value: hex.EncodeToString(res)}

			return cliCtx.PrintOutput(out)
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"VM address (Bech32 / HEX string)",
		"VM path (HEX string)",
	})

	return cmd
}

// Compile returns query command that compiles Move script / module.
func Compile(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "compile [moveFile] [account]",
		Short:   "Compile script / module using source code from Move file",
		Example: "compile script.move wallet196udj7s83uaw2u4safcrvgyqc0sc3flxuherp6 --to-file script.move.json",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			compilerAddr := viper.GetString(vm_client.FlagCompilerAddr)

			// parse inputs
			moveContent, err := helpers.ParseFilePath("moveFile", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			address, err := helpers.ParseSdkAddressParam("account", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare request
			sourceFile := &vm_grpc.SourceFile{
				Text:    string(moveContent),
				Address: common_vm.Bech32ToLibra(address),
			}

			// compile Move file
			bytecode, err := vm_client.Compile(compilerAddr, sourceFile)
			if err != nil {
				return err
			}

			if err := saveOutput(bytecode, cdc); err != nil {
				return fmt.Errorf("error during compiled bytes print: %v", err)
			}
			fmt.Println("Compilation successfully done")

			return nil
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"path to .move file",
		"account address (Bech32 / HEX string)",
	})

	return cmd
}

// GetTxVMStatus returns query command that returns transaction VM status.
func GetTxVMStatus(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tx [hash]",
		Short:   "Get TX VM status by hash",
		Example: "query tx 6D5A4D889BCDB4C71C6AE5836CD8BC1FD8E0703F1580B9812990431D1796CE34",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			output, err := utils.QueryTx(cliCtx, args[0])
			if err != nil {
				return err
			}

			if output.Empty() {
				return fmt.Errorf("no transaction found with hash %s", args[0])
			}

			status := types.NewVMStatusFromABCILogs(output)

			return cliCtx.PrintOutput(status)
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"transaction hash code",
	})

	return cmd
}

// saveOutput prints compilation output to stdout or file.
func saveOutput(bytecode []byte, cdc *codec.Codec) error {
	code := hex.EncodeToString(bytecode)
	output := viper.GetString(vm_client.FlagOutput)

	mvFile := vm_client.MoveFile{Code: code}
	mvBytes, err := cdc.MarshalJSONIndent(mvFile, "", "    ")
	if err != nil {
		return err
	}

	if output == "" || output == "stdout" {
		fmt.Println("Compiled code: ")
		fmt.Println(string(mvBytes))
	} else {
		// write to file output
		if err := ioutil.WriteFile(output, mvBytes, 0644); err != nil {
			return err
		}

		fmt.Printf("Result saved to file %s\n", output)
	}

	return nil
}
