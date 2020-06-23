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
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govCli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	codec "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/cmd/config"
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
		cmd.Flags().String(vmClient.FlagCompilerAddr, config.DefaultCompilerAddr, vmClient.FlagCompilerUsage)
		txCmd.AddCommand(cmd)
	}

	commands := sdkClient.PostCommands(
		DeployContract(cdc),
		flags.LineBreak,
		UpdateStdlibProposal(cdc),
	)
	commands = append(commands, compileCommands...)

	txCmd.AddCommand(commands...)

	return txCmd
}

// Read Move file contains code in hex.
func GetMVFromFile(filePath string) (vmClient.MoveFile, error) {
	var move vmClient.MoveFile

	file, err := os.Open(filePath)
	if err != nil {
		return move, err
	}
	defer file.Close()

	jsonContent, err := ioutil.ReadAll(file)
	if err != nil {
		return move, err
	}

	if err := json.Unmarshal(jsonContent, &move); err != nil {
		return move, err
	}

	return move, nil
}

// Execute script contract.
func ExecuteScript(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "execute [compiledMoveFile] [arg1,arg2,arg3,...] --from [account] --fees [dfiFee] --gas [gas]",
		Short:   "execute Move script",
		Example: "execute ./script.move.json wallet1jk4ld0uu6wdrj9t8u3gghm9jt583hxx7xp7he8 100 true \"my string\" \"68656c6c6f2c20776f726c6421\" #\"DFI_ETH\" --from my_account --fees 1dfi --gas 500000",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			compilerAddr := viper.GetString(vmClient.FlagCompilerAddr)

			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := txBldrCtx.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := cliBldrCtx.NewCLIContextWithInput(inBuf).WithCodec(cdc)
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return fmt.Errorf("provide correct parameter for --from flag: %v", err)
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

			if len(extractedArgs) < len(parsedArgs) {
				return fmt.Errorf("arguments amount is not enough to call script, too many arguments, expected %d", len(extractedArgs))
			}

			if len(extractedArgs) > len(parsedArgs) {
				return fmt.Errorf("arguments amount is not enough to call script, too few arguments, expected %d", len(extractedArgs))
			}

			for i, arg := range parsedArgs {
				switch extractedArgs[i] {
				case vm_grpc.VMTypeTag_Vector:
					// trying to parse hex string
					_, err := hex.DecodeString(arg)
					if err != nil {
						// if failed, convert string to hex
						scriptArgs[i] = types.NewScriptArg(hex.EncodeToString([]byte(arg)), extractedArgs[i])
					} else {
						// otherwise use hex
						scriptArgs[i] = types.NewScriptArg(arg, extractedArgs[i])
					}

				case vm_grpc.VMTypeTag_U8, vm_grpc.VMTypeTag_U64, vm_grpc.VMTypeTag_U128:
					if arg[0] == '#' {
						// try to convert to xxhash
						seed := xxhash.NewS64(0)

						if len(arg) < 2 {
							return fmt.Errorf("incorrect format for xxHash argument (prefixed #) %q", arg)
						}

						_, err := seed.WriteString(strings.ToLower(arg[1:]))
						if err != nil {
							return fmt.Errorf("can't format to xxHash argument %q (format happens because of '#' prefix)", arg)
						}

						arg = strconv.FormatUint(seed.Sum64(), 10)
					}

					n, isOk := sdk.NewIntFromString(arg)
					if !isOk {
						return fmt.Errorf("%s is not a unsigned number (max is unsigned 256), wrong argument type, must be: %s", arg, types.VMTypeTagToStringPanic(extractedArgs[i]))
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
					if _, err := sdk.AccAddressFromBech32(arg); err != nil {
						return fmt.Errorf("can't parse address argument %s, check address and try again: %s", arg, err.Error())
					}

					scriptArgs[i] = types.NewScriptArg(arg, extractedArgs[i])

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
		Use:     "publish [compiledMoveFile] --from [account] --fees [dfiFee] --gas [gas]",
		Short:   "publish Move module",
		Example: "publish ./my_module.move.json --from my_account --fees 1dfi --gas 500000",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := txBldrCtx.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := cliBldrCtx.NewCLIContextWithInput(inBuf).WithCodec(cdc)
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return fmt.Errorf("provide correct parameter for --from flag: %v", err)
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

// Send governance update stdlib VM module proposal.
func UpdateStdlibProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-stdlib-proposal [mvFile] [plannedBlockHeight] [sourceUrl] [updateDescription] [flags]",
		Args:    cobra.ExactArgs(4),
		Short:   "Submit a DVM stdlib update proposal",
		Example: "update-stdlib-proposal ./update.move.json 1000 http://github.com/repo 'fix for Foo module' --deposit 100dfi --from my_account --fees 1dfi",
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := cliBldrCtx.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			// parse inputs
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)
			fromAddress := cliCtx.FromAddress
			if err := accGetter.EnsureExists(fromAddress); err != nil {
				return fmt.Errorf("%s flag: %v", flags.FlagFrom, err)
			}

			depositStr, err := cmd.Flags().GetString(govCli.FlagDeposit)
			if err != nil {
				return fmt.Errorf("%s flag: %w", govCli.FlagDeposit, err)
			}
			deposit, err := sdk.ParseCoins(depositStr)
			if err != nil {
				return fmt.Errorf("%s flag %q: parsing: %w", govCli.FlagDeposit, depositStr, err)
			}

			mvFilePath := args[0]
			mvFile, err := GetMVFromFile(mvFilePath)
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "mvFile", mvFilePath, err)
			}
			code, err := hex.DecodeString(mvFile.Code)
			if err != nil {
				return fmt.Errorf("%s argument %q: decoding: %w", "mvFile", mvFilePath, err)
			}

			plannedBlockHeightStr := args[1]
			plannedBlockHeight, err := strconv.ParseInt(plannedBlockHeightStr, 10, 64)
			if err != nil {
				return fmt.Errorf("%s argument %q: decoding: %w", "plannedBlockHeight", plannedBlockHeightStr, err)
			}

			sourceUrl, updateDesc := args[2], args[3]

			// prepare and send message
			content := types.NewStdlibUpdateProposal(types.NewPlan(plannedBlockHeight), sourceUrl, updateDesc, code)
			if err := content.ValidateBasic(); err != nil {
				return err
			}

			msg := gov.NewMsgSubmitProposal(content, deposit, fromAddress)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(govCli.FlagDeposit, "", "deposit of proposal")

	return cmd
}
