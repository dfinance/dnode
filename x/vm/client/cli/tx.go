package cli

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/OneOfOne/xxhash"
	cliBldrCtx "github.com/cosmos/cosmos-sdk/client/context"
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	codec "github.com/tendermint/go-amino"

	vmClient "github.com/dfinance/dnode/x/vm/client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "VM transactions commands",
	}

	compileCommands := sdkClient.PostCommands(
		ExecuteScript(cdc),
	)
	for _, cmd := range compileCommands {
		cmd.Flags().String(vmClient.FlagCompilerAddr, vmClient.DefaultCompilerAddr, vmClient.FlagCompilerUsage)
		txCmd.AddCommand(cmd)
	}

	commands := sdkClient.PostCommands(DeployContract(cdc))
	commands = append(commands, compileCommands...)

	txCmd.AddCommand(commands...)

	return txCmd
}

// Read MVir file contains code in hex.
func GetMVFromFile(filePath string) (vmClient.MVFile, error) {
	var mvir vmClient.MVFile

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
		Use:     "execute-script [compileMvir] [arg1,arg2,arg3,...]",
		Short:   "execute Move script",
		Example: "execute-script ./script.mvir.json wallet1jk4ld0uu6wdrj9t8u3gghm9jt583hxx7xp7he8 100 true \"my string\" \"68656c6c6f2c20776f726c6421\" #\"DFI_ETH\"",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			compilerAddr := viper.GetString(vmClient.FlagCompilerAddr)

			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := txBldrCtx.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := cliBldrCtx.NewCLIContextWithInput(inBuf).WithCodec(cdc)
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return fmt.Errorf("fromAddress: %w", err)
			}

			mvFile, err := GetMVFromFile(args[0])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "mvFile", args[0], err)
			}

			code, err := hex.DecodeString(mvFile.Code)
			if err != nil {
				return err
			}

			// parsing arguments
			parsedArgs := args[1:]
			scriptArgs := make([]types.ScriptArg, len(parsedArgs))
			extractedArgs, err := vmClient.ExtractArguments(compilerAddr, code)
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
					// trying to parse hex
					_, err := hex.DecodeString(arg)
					if err != nil {
						// if not success, just convert string to hex.
						scriptArgs[i] = types.NewScriptArg(fmt.Sprintf("b\"%s\"", hex.EncodeToString([]byte(arg))), extractedArgs[i])
					} else {
						// otherwise just use hex.
						scriptArgs[i] = types.NewScriptArg(fmt.Sprintf("b\"%s\"", arg), extractedArgs[i])
					}

				case vm_grpc.VMTypeTag_Struct:
					return fmt.Errorf("currently doesnt's support struct type as argument")

				case vm_grpc.VMTypeTag_U8, vm_grpc.VMTypeTag_U64, vm_grpc.VMTypeTag_U128:
					if arg[0] == '#' {
						// try to convert to xxhash
						seed := xxhash.NewS64(0)

						if len(arg) < 2 {
							return fmt.Errorf("incorrect format for xxHash argument (prefixed #) %q", arg)
						}

						fmt.Printf("Result: %s\n", strings.ToLower(arg[1:]))
						_, err := seed.WriteString(strings.ToLower(arg[1:]))
						if err != nil {
							return fmt.Errorf("can't format to xxHash argument %q (format happens because of '#' prefix)", arg)
						}

						arg = strconv.FormatUint(seed.Sum64(), 10)
					}

					n, isOk := sdk.NewIntFromString(arg)

					if !isOk {
						return fmt.Errorf("%s is not a unsigned number (max is unsigned 256), wrong argument type, must be: %s", arg, types.VMTypeToStringPanic(extractedArgs[i]))
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
					addrBech, err := sdk.AccAddressFromBech32(arg)
					if err != nil {
						return fmt.Errorf("can't parse address argument %s, check address and try again: %s", arg, err.Error())
					}

					addr := "0x" + hex.EncodeToString(types.Bech32ToLibra(addrBech))
					scriptArgs[i] = types.NewScriptArg(addr, extractedArgs[i])

				case vm_grpc.VMTypeTag_Bool:
					if arg != "true" && arg != "false" {
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
		Use:     "deploy-module [mvFile]",
		Short:   "deploy Move contract",
		Example: "deploy-module ./my_module.mvir.json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := txBldrCtx.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := cliBldrCtx.NewCLIContextWithInput(inBuf).WithCodec(cdc)
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return fmt.Errorf("fromAddress: %w", err)
			}

			mvFile, err := GetMVFromFile(args[0])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "mvFile", args[0], err)
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
