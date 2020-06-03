package cli

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/common_vm"
	vmClient "github.com/dfinance/dnode/x/vm/client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// Returns get commands for this module.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "VM query commands, includes compiler",
	}

	compileCommands := sdkClient.GetCommands(
		CompileScript(cdc),
		CompileModule(cdc),
	)
	for _, cmd := range compileCommands {
		cmd.Flags().String(vmClient.FlagCompilerAddr, vmClient.DefaultCompilerAddr, vmClient.FlagCompilerUsage)
		cmd.Flags().String(vmClient.FlagOutput, "", "--to-file ./compiled.mv")
	}

	commands := sdkClient.GetCommands(
		GetData(types.ModuleName, cdc),
		GetTransactioVMError(cdc),
	)
	commands = append(commands, compileCommands...)

	queryCmd.AddCommand(commands...)

	return queryCmd
}

// Read move file by file path.
func readMoveFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

// Save output to stdout or file after compilation.
func saveOutput(bytecode []byte, cdc *codec.Codec) error {
	code := hex.EncodeToString(bytecode)
	output := viper.GetString(vmClient.FlagOutput)

	mvFile := vmClient.MoveFile{Code: code}
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

// Get data from data source by access path.
func GetData(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "get-data [address] [path]",
		Short:   "get-data from data source storage by address and path, address could be bech32 or hex",
		Example: "get-data wallet1jk4ld0uu6wdrj9t8u3gghm9jt583hxx7xp7he8 0019b01c2cf3c2160a43e4dcad70e3e5d18151cc38de7a1d1067c6031bfa0ae4d9",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// extract data
			rawAddress := args[0]
			var address sdk.AccAddress
			address, err := hex.DecodeString(rawAddress)
			if err != nil {
				address, err = sdk.AccAddressFromBech32(rawAddress)
				if err != nil {
					return fmt.Errorf("can't parse address: %s\n, check address format, it could be libra hex or bech32", rawAddress)
				}

				address = common_vm.Bech32ToLibra(address)
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

// Compile Move script.
func CompileScript(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "compile-script [moveFile] [account]",
		Short:   "compile script using source code from Move file",
		Example: "compile-script script.move wallet196udj7s83uaw2u4safcrvgyqc0sc3flxuherp6 --to-file script.move.json",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			compilerAddr := viper.GetString(vmClient.FlagCompilerAddr)

			// read provided file
			moveContent, err := readMoveFile(args[0])
			if err != nil {
				return fmt.Errorf("error during reading Move file %q: %v", args[0], err)
			}

			addr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return fmt.Errorf("error during parsing address %s: %v", args[1], err)
			}

			// Move file
			sourceFile := &vm_grpc.MvIrSourceFile{
				Text:    string(moveContent),
				Address: common_vm.Bech32ToLibra(addr),
				Type:    vm_grpc.ContractType_Script,
			}

			// compile Move file
			bytecode, err := vmClient.Compile(compilerAddr, sourceFile)
			if err != nil {
				return err
			}

			if err := saveOutput(bytecode, cdc); err != nil {
				return fmt.Errorf("error during compiled bytes output: %v", err)
			}

			fmt.Println("Compilation successful done.")

			return nil
		},
	}
}

// Compile Move module.
func CompileModule(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "compile-module [moveFile] [account]",
		Short:   "compile module connected to account, using source code from Move file",
		Example: "compile-module module.move wallet196udj7s83uaw2u4safcrvgyqc0sc3flxuherp6 --to-file module.move.json",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			compilerAddr := viper.GetString(vmClient.FlagCompilerAddr)

			// read provided file
			moveContent, err := readMoveFile(args[0])
			if err != nil {
				return fmt.Errorf("error during reading Move file %q: %v", args[0], err)
			}

			addr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return fmt.Errorf("error during parsing address %s: %v", args[1], err)
			}

			// Move file
			sourceFile := &vm_grpc.MvIrSourceFile{
				Text:    string(moveContent),
				Address: common_vm.Bech32ToLibra(addr),
				Type:    vm_grpc.ContractType_Module,
			}

			// compile Move file
			bytecode, err := vmClient.Compile(compilerAddr, sourceFile)
			if err != nil {
				return err
			}

			if err := saveOutput(bytecode, cdc); err != nil {
				return fmt.Errorf("error during compiled bytes output: %v", err)
			}

			fmt.Println("Compilation successful done.")

			return nil
		},
	}
}

// Get transaction VM errors if contains it.
func GetTransactioVMError(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "tx [hash]",
		Short:   "query tx vm error by hash",
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

			isFound, errResp := types.NewVMErrorFromABCILogs(output)

			if !isFound {
				return fmt.Errorf("transaction %s doesn't contain vm errors", args[0])
			}

			return cliCtx.PrintOutput(errResp)
		},
	}
}
