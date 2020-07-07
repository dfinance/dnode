package cli

import (
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

// GetCmdPostOrder returns tx command which post a new order.
func GetCmdPostOrder(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "post [market_id] [direction] [price] [quantity] [TTL_in_sec]",
		Example: "post 0 bid 100 100000000 --from wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m",
		Short:   "Post a new order",
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			marketID, err := helpers.ParseDnIDParam("market_id", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			direction := types.Direction(strings.ToLower(args[1]))
			if !direction.IsValid() {
				return helpers.BuildError("direction", args[1], helpers.ParamTypeCliArg, "invalid (bid / ask)")
			}

			price, err := helpers.ParseSdkUintParam("price", args[2], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			quantity, err := helpers.ParseSdkUintParam("quantity", args[3], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			ttlInSec, err := helpers.ParseUint64Param("TTL_in_sec", args[4], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send message
			msg := types.NewMsgPost(fromAddr, marketID, direction, price, quantity, ttlInSec)

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"market ID [uint]",
		"order type [bid/ask]",
		"price with decimals (1.0 BTC with 8 decimals -> 100000000)",
		"quantity with decimals",
		"order TTL [s]",
	})

	return cmd
}

// GetCmdRevokeOrder returns tx command which revokes an order.
func GetCmdRevokeOrder(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "revoke [order_id]",
		Short:   "Revoke an order",
		Example: "revoke 0 --from wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			orderID, err := helpers.ParseDnIDParam("order_id", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send message
			msg := types.NewMsgRevokeOrder(fromAddr, orderID)

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"order ID [uint]",
	})

	return cmd
}
