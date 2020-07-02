package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govCli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// PostWithdrawCurrency returns tx command which post a new withdraw request.
func PostWithdrawCurrency(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraw-currency [denom] [amount] [pegZoneSpender] [pegZoneChainID]",
		Short:   "Withdraw issued currency, reducing spender balance",
		Example: "withdraw-currency dfi 100 {account} testnet --from {account}",
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			amount, err := helpers.ParseSdkIntParam("amount", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send message
			msg := types.NewMsgWithdrawCurrency(args[0], amount, fromAddr, args[2], args[3])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"currency denomination symbol",
		"reduce coin amount",
		"spender address for PegZone",
		"chainID for PegZone blockchain",
	})

	return cmd
}

// Send governance add currency proposal.
func AddCurrencyProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-currency-proposal [denom] [decimals] [vmBalancePathHex] [vmInfoPathHex]",
		Args:    cobra.ExactArgs(4),
		Short:   "Submit currency add proposal, creating non-token currency",
		Example: "add-currency-proposal dfi 18 {balancePath} {infoPath} --deposit 100dfi --fees 1dfi",
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

			denom := args[0]
			if err := helpers.ValidateDenomParam("denom", denom, helpers.ParamTypeCliArg); err != nil {
				return err
			}

			decimals, err := helpers.ParseUint8Param("decimals", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			balancePath, err := helpers.ParseHexStringParam("vmBalancePathHex", args[2], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			infoPath, err := helpers.ParseHexStringParam("vmInfoPathHex", args[3], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send message
			content := types.NewAddCurrencyProposal(denom, decimals, balancePath, infoPath)
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
	cmd.Flags().String(govCli.FlagDeposit, "", "deposit of proposal")
	helpers.BuildCmdHelp(cmd, []string{
		"new currency denomination symbol",
		"new currency number of decimals",
		"DVM path for balance resources [HEX string]",
		"DVM path for currencyInfo resource [HEX string]",
	})

	return cmd
}
