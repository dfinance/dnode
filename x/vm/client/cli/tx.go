package cli

import (
	"encoding/hex"
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	cliBldrCtx "github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
	codec "github.com/tendermint/go-amino"
	"io/ioutil"
	"os"
	"wings-blockchain/x/vm/internal/types"
)

// MVir struct contains code from file in hex.
type MVir struct {
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
func GetMVirFromFile(filePath string) (MVir, error) {
	var mvir MVir

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
		Use:   "execute-script [fileVmir]",
		Short: "execute Move script",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := cliBldrCtx.NewCLIContext().WithCodec(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}

			mvir, err := GetMVirFromFile(args[0])
			if err != nil {
				return err
			}

			code, err := hex.DecodeString(mvir.Code)
			if err != nil {
				return err
			}

			msg := types.NewMsgScriptContract(cliCtx.GetFromAddress(), code)
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
		Use:   "deploy [fileMvir]",
		Short: "deploy Move contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := cliBldrCtx.NewCLIContext().WithCodec(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}

			mvir, err := GetMVirFromFile(args[0])
			if err != nil {
				return err
			}

			code, err := hex.DecodeString(mvir.Code)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeployContract(cliCtx.GetFromAddress(), code)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
