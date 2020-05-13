package cli

import (
	"bufio"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/x/market/internal/types"
)

// GetCmdAddMarket returns tx command which adds a market object.
func GetCmdAddMarket(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "add [base_denom] [quote_denom]",
		Example: "dncli market add dfi eth",
		Short:   "Add a new market",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			// message send
			msg := types.NewMsgCreateMarket(args[0], args[1])

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
