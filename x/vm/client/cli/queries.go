package cli

import (
	connContext "context"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc"
	"io/ioutil"
	"os"
	vmConfig "wings-blockchain/cmd/config"
	"wings-blockchain/x/vm/internal/types"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

const (
	FlagOutput = "to-file"
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
		cmd.Flags().String(FlagOutput, "", "--to-file ./compiled.mv")
	}

	commands := client.GetCommands(
		GetData("vm", cdc),
	)
	commands = append(commands, compileCommands...)

	queries.AddCommand(
		commands...,
	)

	return queries
}

// Create connection to virtual machine.
func createVMConn(config *vmConfig.VMConfig) (*grpc.ClientConn, error) {
	return grpc.Dial(config.Address, grpc.WithInsecure())
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

func compile(config *vmConfig.VMConfig, sourceFile *vm_grpc.MvIrSourceFile) ([]byte, bool) {
	conn, err := createVMConn(config)
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
			fmt.Printf("Error from compiler: %s\n", string(err))
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
func CompileScript(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "compile_script [mvirFile]",
		Short: "compile script using source code from mvir file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := vmConfig.ReadVMConfig(viper.GetString(cli.HomeFlag))
			if err != nil {
				fmt.Println("Error during reading config.")
				return err
			}

			// read provided file
			mvirContent, err := readMvirFile(args[0])
			if err != nil {
				fmt.Println("Error during reading mvir file.")
				return err
			}

			// Mvir file
			sourceFile := &vm_grpc.MvIrSourceFile{
				Text:    string(mvirContent),
				Address: nil,
				Type:    vm_grpc.ContractType_Script,
			}

			fmt.Printf("source code: %s", string(mvirContent))

			// compile mvir file
			bytecode, isOk := compile(config, sourceFile)
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
		Use:   "compile_module [mvirFile] [account]",
		Short: "compile module connected to account, using source code from mvir file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := vmConfig.ReadVMConfig(viper.GetString(cli.HomeFlag))
			if err != nil {
				fmt.Println("Error during reading config.")
				return err
			}

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
			bytecode, isOk := compile(config, sourceFile)
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
