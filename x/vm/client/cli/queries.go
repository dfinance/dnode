package cli

import (
	connContext "context"
	"encoding/hex"
	"fmt"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types/vm_grpc"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"io/ioutil"
	"os"
)

const (
	FlagOutput       = "to-file"
	FlagCompilerAddr = "compiler"
)

// Get query commands for VM module.
func GetQueriesCmd(cdc *codec.Codec) *cobra.Command {
	queries := &cobra.Command{
		Use:   "vm",
		Short: "VM query commands, include compiler",
	}

	compileCommands := client.GetCommands(
		CompileScript(cdc),
		CompileModule(cdc),
	)

	for _, cmd := range compileCommands {
		cmd.Flags().String(FlagCompilerAddr, "127.0.0.1:50053", "--compiler 127.0.0.1:50053")
		cmd.Flags().String(FlagOutput, "", "--to-file ./compiled.mv")
	}

	commands := client.GetCommands(
		GetData("vm", cdc),
		GetDenomHex(cdc),
	)
	commands = append(commands, compileCommands...)

	queries.AddCommand(
		commands...,
	)

	return queries
}

// Create connection to virtual machine.
func createVMConn() (*grpc.ClientConn, error) {
	return grpc.Dial(viper.GetString(FlagCompilerAddr), grpc.WithInsecure())
}

// Read mvir file by file path.
func readMvirFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

func compile(sourceFile *vm_grpc.MvIrSourceFile) ([]byte, bool) {
	conn, err := createVMConn()
	if err != nil {
		fmt.Printf("Compilation failed because of error during connection to VM: %s\n", err.Error())
		return nil, false
	}
	defer conn.Close()

	client := vm_grpc.NewVMCompilerClient(conn)
	connCtx := connContext.Background()

	resp, err := client.Compile(connCtx, sourceFile)
	if err != nil {
		fmt.Printf("Compilation failed because of error during compilation and connection to VM: %s\n", err.Error())
		return nil, false
	}

	// if contains errors
	if len(resp.Errors) > 0 {
		for _, err := range resp.Errors {
			fmt.Printf("Error from compiler: %s\n", err)
		}
		fmt.Println("Compilation failed because of errors from compiler.")
		return nil, false
	}

	return resp.Bytecode, true
}

// Save output to stdout or file after compilation.
func saveOutput(bytecode []byte, cdc *codec.Codec) error {
	code := hex.EncodeToString(bytecode)
	output := viper.GetString(FlagOutput)

	mvFile := MVFile{Code: code}
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
		Example: "get-data wallets196udj7s8.../00000000... 0019b01c2cf3c2160a43e4dcad70e3e5d18151cc38de7a1d1067c6031bfa0ae4d9",
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
					fmt.Printf("can't parse address: %s\n, check address format, it could be libra hex or bech32\n", rawAddress)
					return nil
				}

				address, err = hex.DecodeString(types.Bech32ToLibra(address))
				if err != nil {
					return err
				}
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

func GetDenomHex(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "denom-hex [denom]",
		Short: "get denom in hex representation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hex := hex.EncodeToString([]byte(args[0]))
			fmt.Printf("Denom in hex: %s\n", hex)
			return nil
		},
	}
}

// Compile Mvir script.
func CompileScript(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "compile-script [mvirFile] [account]",
		Short:   "compile script using source code from mvir file",
		Example: "compile-script script.mvir wallets196udj7s83uaw2u4safcrvgyqc0sc3flxuherp6:Address --to-file script.mv --compiler 127.0.0.1:50053",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// read provided file
			mvirContent, err := readMvirFile(args[0])
			if err != nil {
				fmt.Println("Error during reading mvir file.")
				return err
			}

			// Mvir file
			sourceFile := &vm_grpc.MvIrSourceFile{
				Text:    string(mvirContent),
				Address: []byte(args[1]),
				Type:    vm_grpc.ContractType_Script,
			}

			// compile mvir file
			bytecode, isOk := compile(sourceFile)
			if !isOk {
				return nil
			}

			err = saveOutput(bytecode, cdc)
			if err != nil {
				fmt.Println("Error during compiled bytes output.")
				return err
			}

			fmt.Println("Compilation successful done.")

			return nil
		},
	}
}

// Compile Mvir module.
func CompileModule(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "compile-module [mvirFile] [account]",
		Short:   "compile module connected to account, using source code from mvir file",
		Example: "compile-module module.mvir wallets196udj7s83uaw2u4safcrvgyqc0sc3flxuherp6:Address --to-file module.mv --compiler 127.0.0.1:50053",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// read provided file
			mvirContent, err := readMvirFile(args[0])
			if err != nil {
				fmt.Println("Error during reading mvir file.")
				return err
			}

			// Mvir file
			sourceFile := &vm_grpc.MvIrSourceFile{
				Text:    string(mvirContent),
				Address: []byte(args[1]),
				Type:    vm_grpc.ContractType_Module,
			}

			// compile mvir file
			bytecode, isOk := compile(sourceFile)
			if !isOk {
				return nil
			}

			err = saveOutput(bytecode, cdc)
			if err != nil {
				fmt.Println("Error during compiled bytes output.")
				return err
			}

			fmt.Println("Compilation successful done.")

			return nil
		},
	}
}
