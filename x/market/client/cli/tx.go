package cli

import (
	"bufio"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/x/market/internal/types"
)

// GetCmdAddMarket return tx command which adds a market object.
func GetCmdAddMarket(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "add [base_denom] [quote_denom] [base_decimals]",
		Example: "dncli market add dfi eth --from wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m",
		Short:   "Add a new market",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			// inputs parsing
			baseDecimals, err := strconv.ParseUint(args[2], 10, 8)
			if err != nil {
				return fmt.Errorf("%s argument %q: parsing uint: %w", "base_decimals", args[2], err)
			}

			// message send
			msg := types.NewMsgCreateMarket(cliCtx.GetFromAddress(), args[0], args[1], uint8(baseDecimals))

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
