package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govCli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	codec "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/vm/client/vm_client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// ExecuteScript returns tx command which executed VM script.
func ExecuteScript(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "execute [moveFile] [arg1,arg2,arg3,..]",
		Short:   "Execute Move script",
		Example: "execute ./script.move.json wallet1jk4ld0uu6wdrj9t8u3gghm9jt583hxx7xp7he8 100 true \"my string\" \"68656c6c6f2c20776f726c6421\" #\"DFI_ETH\" --from my_account --fees 10000dfi --gas 500000",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())
			compilerAddr := viper.GetString(vm_client.FlagCompilerAddr)

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			code, err := getMoveCodeFromFileArg(args[0])
			if err != nil {
				return err
			}

			strArgs := args[1:]
			typedArgs, err := vm_client.ExtractArguments(compilerAddr, code)
			if err != nil {
				return fmt.Errorf("extracting typed args from the code: %w", err)
			}

			scriptArgs, err := vm_client.ConvertStringScriptArguments(strArgs, typedArgs)
			if err != nil {
				return fmt.Errorf("converting input args to typed args: %w", err)
			}
			if len(scriptArgs) == 0 {
				scriptArgs = nil
			}

			// prepare and send message
			msg := types.NewMsgExecuteScript(fromAddr, code, scriptArgs)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"path to compiled Mode file containing bytecode",
		"space separated VM script arguments (optional)",
	})

	return cmd
}

// DeployContract returns tx command which deploys VM module (contract).
func DeployContract(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "publish [moveFile]",
		Short:   "Publish Move module",
		Example: "publish ./my_module.move.json --from my_account --fees 10000dfi --gas 500000",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			code, err := getMoveCodeFromFileArg(args[0])
			if err != nil {
				return err
			}

			// prepare and send message
			msg := types.NewMsgDeployModule(fromAddr, code)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"path to compiled Mode file containing bytecode",
	})

	return cmd
}

// UpdateStdlibProposal returns tx command which sends governance update stdlib VM module proposal.
func UpdateStdlibProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-stdlib-proposal [moveFile] [plannedBlockHeight] [sourceUrl] [updateDescription]",
		Short:   "Submit a DVM stdlib update proposal",
		Example: "update-stdlib-proposal ./update.move.json 1000 http://github.com/repo 'fix for Foo module' --deposit 10000dfi --from my_account --fees 10000dfi",
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			deposit, err := helpers.ParseDepositFlag(cmd.Flags())
			if err != nil {
				return err
			}

			code, err := getMoveCodeFromFileArg(args[0])
			if err != nil {
				return err
			}

			plannedBlockHeight, err := helpers.ParseInt64Param("plannedBlockHeight", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			sourceUrl, updateDesc := args[2], args[3]

			// prepare and send message
			content := types.NewStdlibUpdateProposal(types.NewPlan(plannedBlockHeight), sourceUrl, updateDesc, code)
			if err := content.ValidateBasic(); err != nil {
				return err
			}

			msg := gov.NewMsgSubmitProposal(content, deposit, fromAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"path to compiled Mode file containing bytecode",
		"blockHeight at which update should occur [int]",
		"URL containing proposal source code",
		"proposal description (version, short changelist)",
	})
	cmd.Flags().String(govCli.FlagDeposit, "", "deposit of proposal")

	return cmd
}

// getMoveCodeFromFileArg reads .move file and converts its code field.
func getMoveCodeFromFileArg(argValue string) (moveCode []byte, retErr error) {
	const argName = "moveFile"

	jsonContent, err := helpers.ParseFilePath(argName, argValue, helpers.ParamTypeCliArg)
	if err != nil {
		retErr = err
		return
	}

	var moveFile vm_client.MoveFile
	if err := json.Unmarshal(jsonContent, &moveFile); err != nil {
		retErr = helpers.BuildError(
			argName,
			argValue,
			helpers.ParamTypeCliArg,
			fmt.Sprintf("Move file JSON unmarshal: %v", err),
		)
		return
	}

	code, err := hex.DecodeString(moveFile.Code)
	if err != nil {
		retErr = helpers.BuildError(
			argName,
			argValue,
			helpers.ParamTypeCliArg,
			fmt.Sprintf("Move file code HEX decode: %v", err),
		)
		return
	}
	moveCode = code

	return
}
