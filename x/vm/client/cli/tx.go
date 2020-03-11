package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types/vm_grpc"
	"github.com/cosmos/cosmos-sdk/client"
	cliBldrCtx "github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
	codec "github.com/tendermint/go-amino"
	"io/ioutil"
	"os"
)

// Return TX commands for CLI.
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "VM transactions commands",
	}

	compileCommands := client.PostCommands(
		ExecuteScript(cdc),
	)

	for _, cmd := range compileCommands {
		cmd.Flags().String(FlagCompilerAddr, FlagCompilerDefault, FlagCompilerUsage)
		txCmd.AddCommand(cmd)
	}

	txCmd.AddCommand(client.PostCommands(DeployContract(cdc))...)
	txCmd.AddCommand(compileCommands...)

	return txCmd
}

// Read MVir file contains code in hex.
func GetMVFromFile(filePath string) (MVFile, error) {
	var mvir MVFile

	file, err := os.Open(filePath)
	if err != nil {
		return mvir, err
	}
	defer file.Close()

	jsonContent, err := ioutil.ReadAll(file)
	if err != nil {
		return mvir, err
	}

	if err := json.Unmarshal(jsonContent, &mvir); err != nil {
		return mvir, err
	}

	return mvir, nil
}

// Execute script contract.
func ExecuteScript(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "execute-script [mvFile] [arg1:type1,arg2:type2,...]",
		Short: "execute Move script",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := cliBldrCtx.NewCLIContext().WithCodec(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}

			mvFile, err := GetMVFromFile(args[0])
			if err != nil {
				return err
			}

			code, err := hex.DecodeString(mvFile.Code)
			if err != nil {
				return err
			}

			// parsing arguments
			parsedArgs := args[1:]
			scriptArgs := make([]types.ScriptArg, len(parsedArgs))
			extractedArgs, err := extractArgs(code)
			if err != nil {
				return err
			}

			if len(extractedArgs) != len(parsedArgs) {
				// error not enough args
				return fmt.Errorf("arguments amount is not enough to call script, some arguments missed")
			}

			for i, arg := range parsedArgs {
				switch extractedArgs[i] {
				case vm_grpc.VMTypeTag_ByteArray:
					scriptArgs[i] = types.NewScriptArg(fmt.Sprintf("b\"%s\"", hex.EncodeToString([]byte(arg))), extractedArgs[i])

				case vm_grpc.VMTypeTag_Struct:
					return fmt.Errorf("currently doesnt's support struct type as argument")

				case vm_grpc.VMTypeTag_U8, vm_grpc.VMTypeTag_U64, vm_grpc.VMTypeTag_U128:
					n, isOk := sdk.NewIntFromString(arg)

					if !isOk {
						return fmt.Errorf("%s is not a unsigned number, wrong argument type, must be: %s", arg, types.VMTypeToStringPanic(extractedArgs[i]))
					}

					switch extractedArgs[i] {
					case vm_grpc.VMTypeTag_U8:
						if n.BigInt().BitLen() > 8 {
							return fmt.Errorf("argument %s must be U8, current bit length is %d, overflow", arg, n.BigInt().BitLen())
						}

					case vm_grpc.VMTypeTag_U64:
						if n.BigInt().BitLen() > 64 {
							return fmt.Errorf("argument %s must be U64, current bit length is %d, overflow", arg, n.BigInt().BitLen())
						}

					case vm_grpc.VMTypeTag_U128:
						if n.BigInt().BitLen() > 128 {
							return fmt.Errorf("argument %s must be U128, current bit length is %d, overflow", arg, n.BigInt().BitLen())
						}
					}

					scriptArgs[i] = types.NewScriptArg(arg, extractedArgs[i])

				case vm_grpc.VMTypeTag_Address:
					// validate address
					_, err := sdk.AccAddressFromBech32(arg)
					if err != nil {
						return fmt.Errorf("can't parse address argument %s, check address and try again: %s", arg, err.Error())
					}
					scriptArgs[i] = types.NewScriptArg(arg, extractedArgs[i])

				case vm_grpc.VMTypeTag_Bool:
					if arg != "true" || arg != "false" {
						return fmt.Errorf("%s argument must be bool, means \"true\" or \"false\"", arg)
					}
					scriptArgs[i] = types.NewScriptArg(arg, extractedArgs[i])

				default:
					scriptArgs[i] = types.NewScriptArg(arg, extractedArgs[i])
				}
			}

			if len(scriptArgs) == 0 {
				scriptArgs = nil
			}

			msg := types.NewMsgExecuteScript(cliCtx.GetFromAddress(), code, scriptArgs)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// Deploy contract cli TX command.
func DeployContract(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deploy-module [mvFile]",
		Short: "deploy Move contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := cliBldrCtx.NewCLIContext().WithCodec(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}

			mvFile, err := GetMVFromFile(args[0])
			if err != nil {
				return err
			}

			code, err := hex.DecodeString(mvFile.Code)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeployModule(cliCtx.GetFromAddress(), code)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
