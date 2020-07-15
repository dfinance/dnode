package cli

import (
	"bufio"
	"fmt"
	"github.com/dfinance/dnode/helpers"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// GetCmdPostPrice returns tx command for posting price for a particular asset.
func GetCmdPostPrice(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "postprice [nomineeKey] [assetCode] [price] [receivedAt]",
		Example: "dncli oracle postprice wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt 100 1594732456",
		Short:   "post the latest price for a particular asset",
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			rawPrice := args[2]
			price, ok := sdk.NewIntFromString(rawPrice)
			if !ok {
				return fmt.Errorf("%s argument %q: wrong value for price", "price", args[2])
			}
			receivedAtInt, ok := sdk.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("%s argument %q: wrong value for time", "receivedAt", args[3])
			}
			receivedAt := tmtime.Canonical(time.Unix(receivedAtInt.Int64(), 0))
			msg := types.NewMsgPostPrice(cliCtx.GetFromAddress(), assetCode, price, receivedAt)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	helpers.BuildCmdHelp(cmd, []string{
		"nomineeKey [string] nominee key or address",
		"assetCode [string] asset code symbol",
		"price [uint] uint format price",
		"receivedAt [uint] unix timestamp",
	})

	return cmd
}

// GetCmdAddOracle returns tx command for adding new oracle for a particular asset.
func GetCmdAddOracle(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-oracle [nomineeKey] [assetCode] [oracleAddress]",
		Example: "dncli oracle add-oracle wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Short:   "Add a new oracle for a particular asset",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			oracleAddr, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "oracle_address", args[2], err)
			}

			msg := types.NewMsgAddOracle(cliCtx.GetFromAddress(), assetCode, oracleAddr)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	helpers.BuildCmdHelp(cmd, []string{
		"nomineeKey [string] nominee key or address",
		"assetCode [string] asset code symbol",
		"oracleAddresses [string] oracle address",
	})

	return cmd
}

// GetCmdSetOracles returns tx command for sets oracles for a particular asset.
func GetCmdSetOracles(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-oracles [nomineeKey] [assetCode] [oracleAddresses]",
		Example: "dncli oracle set-oracles wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt wallet10ff6y8gm2re6awfwz5dvesar8jq02tx7vcvuxn,wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Short:   "Sets a list of oracles for a particular asset",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			oracles, err := types.ParseOracles(args[2])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "oracle_addresses", args[2], err)
			}

			msg := types.NewMsgSetOracles(cliCtx.GetFromAddress(), assetCode, oracles)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	helpers.BuildCmdHelp(cmd, []string{
		"nomineeKey [string] nominee key or address",
		"assetCode [string] asset code symbol",
		"oracleAddresses [string] comma separated list of oracle addresses",
	})

	return cmd
}

// GetCmdAddAsset returns tx command for add a new asset for a list of oracles.
func GetCmdAddAsset(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-asset [nomineeKey] [assetCode] [oracleAddresses]",
		Example: "dncli oracle add-asset wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Short:   "Add a new asset for a list of oracles",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			oracles, err := types.ParseOracles(args[2])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "oracles", args[2], err)
			}
			if len(oracles) == 0 {
				return fmt.Errorf("%s argument %q: empty slice", "oracles", args[2])
			}

			asset := types.NewAsset(assetCode, oracles, true)
			if err := asset.ValidateBasic(); err != nil {
				return err
			}

			msg := types.NewMsgAddAsset(cliCtx.GetFromAddress(), asset)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	helpers.BuildCmdHelp(cmd, []string{
		"nomineeKey [string] nominee key or address",
		"assetCode [string] asset code symbol",
		"oracleAddresses [string] comma separated list of oracle addresses",
	})

	return cmd
}

// GetCmdSetAsset returns tx command for set the existing asset for a list of oracles.
func GetCmdSetAsset(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-asset [nomineeKey] [assetCode] [oracleAddresses]",
		Example: "dncli oracle set-asset wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Short:   "Set the existing asset for a list of oracles",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			oracles, err := types.ParseOracles(args[2])
			if err != nil {
				return fmt.Errorf("%s argument %q: %v", "oracles", args[2], err)
			}
			if len(oracles) == 0 {
				return fmt.Errorf("%s argument %q: empty slice", "oracles", args[2])
			}

			asset := types.NewAsset(assetCode, oracles, true)
			if err := asset.ValidateBasic(); err != nil {
				return err
			}

			msg := types.NewMsgSetAsset(cliCtx.GetFromAddress(), asset)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	helpers.BuildCmdHelp(cmd, []string{
		"nomineeKey [string] nominee key or address",
		"assetCode [string] asset code symbol",
		"oracleAddresses [string] comma separated list of oracle addresses",
	})

	return cmd
}
