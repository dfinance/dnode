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
		Use:     "withdraw [coin] [pegZonePayee] [pegZoneChainID]",
		Short:   "Withdraw issued currency from dfinance chain to pegZone chain, reducing spender balance",
		Example: "withdraw 100xfi {account} testnet --from {account}",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			coin, err := helpers.ParseCoinParam("coin", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send message
			msg := types.NewMsgWithdrawCurrency(coin, fromAddr, args[1], args[2])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"currency denomination symbol and amount in Coin format (1.0 btc with 8 decimals -> 100000000btc)",
		"payee address for PegZone (whose balance is increased)",
		"chainID for PegZone blockchain",
	})

	return cmd
}

// Send governance add currency proposal.
func AddCurrencyProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-currency-proposal [denom] [decimals] [vmBalancePathHex] [vmInfoPathHex] [ERC20ContractAddress]",
		Args:    cobra.ExactArgs(5),
		Short:   "Submit currency add proposal, creating currency",
		Example: "add-currency-proposal xfi 18 {balancePath} {infoPath} --deposit 100xfi --fees 1xfi",
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

			balancePath, _, err := helpers.ParseHexStringParam("vmBalancePathHex", args[2], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			infoPath, _, err := helpers.ParseHexStringParam("vmInfoPathHex", args[3], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			contractAddress := args[4]
			if contractAddress != "" {
				_, err := helpers.ParseEthereumAddressParam("ERC20ContractAddress", contractAddress, helpers.ParamTypeCliArg)
				if err != nil {
					return err
				}
			}

			// prepare and send message
			content := types.NewAddCurrencyProposal(denom, decimals, balancePath, infoPath, contractAddress)
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
		"ERC20 contract address",
	})

	return cmd
}
