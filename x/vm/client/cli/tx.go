package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types"
	"github.com/cosmos/cosmos-sdk/client"
	cliBldrCtx "github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
	codec "github.com/tendermint/go-amino"
	"io/ioutil"
	"os"
	"strings"
)

// MVFile struct contains code from file in hex.
type MVFile struct {
	Code string `json:"code"`
}

// Return TX commands for CLI.
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "VM transactions commands",
	}

	txCmd.AddCommand(client.PostCommands(
		DeployContract(cdc),
		ExecuteScript(cdc),
	)...)

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

			for i, pArg := range parsedArgs {
				parts := strings.Split(pArg, ":")

				if len(parts) != 2 {
					return fmt.Errorf("can't parse argument: %s, check correctness of value and type, also seperator format", pArg)
				}

				typeTag, err := types.GetVMTypeByString(parts[1])
				if err != nil {
					return err
				}

				if len(parts) > 0 {
					scriptArgs[i] = types.NewScriptArg(parts[0], typeTag)
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
