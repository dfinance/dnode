package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

// GetCmdPostOrder returns tx command which post a new order.
func GetCmdPostOrder(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "post [market_id] [direction] [price] [quantity] [TTL_in_sec]",
		Example: "dncli orders post 0 bid 100 100000000 --from wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m",
		Short:   "Post a new order",
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)
			accGetter := auth.NewAccountRetriever(cliCtx)

			// parse inputs
			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return fmt.Errorf("fromAddress: %w", err)
			}

			marketID, err := dnTypes.NewIDFromString(args[0])
			if err != nil {
				return fmt.Errorf("%s argument %q parse error: %w", "market_id", args[0], err)
			}

			direction := types.Direction(strings.ToLower(args[1]))
			if !direction.IsValid() {
				return fmt.Errorf("argument %q: invalid (bid / ask)", "direction")
			}

			price, err := sdk.ParseUint(args[2])
			if err != nil {
				return fmt.Errorf("argument %q: parsing uint: %w", "price", err)
			}

			quantity, err := sdk.ParseUint(args[3])
			if err != nil {
				return fmt.Errorf("argument %q: parsing uint: %w", "quantity", err)
			}

			ttlInSec, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return fmt.Errorf("argument %q: parsing uint: %w", "TTL_in_sec", err)
			}

			// prepare and send message
			msg := types.NewMsgPost(cliCtx.GetFromAddress(), marketID, direction, price, quantity, ttlInSec)

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdRevokeOrder returns tx command which revokes an order.
func GetCmdRevokeOrder(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "revoke [order-id]",
		Short: "Revoke an order",
		Example: "dncli orders revoke 0 --from wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)
			accGetter := auth.NewAccountRetriever(cliCtx)

			// parse inputs
			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return fmt.Errorf("fromAddress: %w", err)
			}

			orderID, err := dnTypes.NewIDFromString(args[0])
			if err != nil {
				return fmt.Errorf("%s argument %q parse error: %w", "order", args[0], err)
			}

			// prepare and send message
			msg := types.NewMsgRevokeOrder(cliCtx.GetFromAddress(), orderID)

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
