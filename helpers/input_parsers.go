package helpers

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	cliCtx "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

type ParamType string

const (
	ParamTypeCliFlag   ParamType = "flag"
	ParamTypeCliArg    ParamType = "argument"
	ParamTypeRestQuery ParamType = "query param"
)

func GetTxCmdCtx(cdc *codec.Codec, cmdInputBuf io.Reader) (cliCtx cliCtx.CLIContext, txBuilder authTypes.TxBuilder) {
	inBuf := bufio.NewReader(cmdInputBuf)
	cliCtx = context.NewCLIContextWithInput(inBuf).WithCodec(cdc)
	txBuilder = authTypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

	return
}

func ParseFromFlag(cliCtx cliCtx.CLIContext) (sdk.AccAddress, error) {
	accGetter := authTypes.NewAccountRetriever(cliCtx)

	if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
		return sdk.AccAddress{}, fmt.Errorf("%s %s: %w", flags.FlagFrom, ParamTypeCliFlag, err)
	}

	return cliCtx.FromAddress, nil
}

func ParsePaginationParams(pageStr, limitStr string, paramType ParamType) (page, limit sdk.Uint, retErr error) {
	parseUint := func(paramName, paramValue string) (sdk.Uint, error) {
		valueInt, ok := sdk.NewIntFromString(paramValue)
		if !ok {
			return sdk.Uint{}, fmt.Errorf("%s %s %q: Int parsing: failed", paramName, paramType, paramValue)
		}
		if valueInt.LT(sdk.OneInt()) {
			return sdk.Uint{}, fmt.Errorf("%s %s %q: Uint parsing: value is less than 1", paramName, paramType, paramValue)
		}
		return sdk.NewUintFromBigInt(valueInt.BigInt()), nil
	}

	if pageStr == "" {
		pageStr = "1"
	}
	if limitStr == "" {
		limitStr = "100"
	}

	page, retErr = parseUint("page", pageStr)
	if retErr != nil {
		return
	}

	limit, retErr = parseUint("limit", limitStr)
	if retErr != nil {
		return
	}

	return
}

func ParseSdkIntParam(argName, argValue string, paramType ParamType) (sdk.Int, error) {
	v, ok := sdk.NewIntFromString(argValue)
	if !ok {
		return sdk.Int{}, fmt.Errorf("%s %s %q: parsing Int: failed", argName, paramType, argValue)
	}

	return v, nil
}

func ParseSdkUintParam(argName, argValue string, paramType ParamType) (sdk.Uint, error) {
	vInt, ok := sdk.NewIntFromString(argValue)
	if !ok {
		return sdk.Uint{}, fmt.Errorf("%s %s %q: parsing Uint: failed", argName, paramType, argValue)
	}

	if vInt.LT(sdk.ZeroInt()) {
		return sdk.Uint{}, fmt.Errorf("%s %s %q: parsing Uint: less than zero", argName, paramType, argValue)
	}

	return sdk.NewUintFromBigInt(vInt.BigInt()), nil
}

func ParseUint8Param(argName, argValue string, paramType ParamType) (uint8, error) {
	v, err := strconv.ParseInt(argValue, 10, 8)
	if err != nil {
		return uint8(0), fmt.Errorf("%s %s %q: uint8 parsing: %w", argName, paramType, argValue, err)
	}

	return uint8(v), nil
}

func ParseSdkAddressParam(argName, argValue string, paramType ParamType) (sdk.AccAddress, error) {
	if v, err := sdk.AccAddressFromBech32(argValue); err == nil {
		return v, nil
	} else if v, err := sdk.AccAddressFromHex(argValue); err == nil {
		return v, nil
	}

	return sdk.AccAddress{}, fmt.Errorf("%s %s %q: parsing Bech32 / HEX account address: failed", argName, paramType, argValue)
}

func ParseDnIDParam(argName, argValue string, paramType ParamType) (dnTypes.ID, error) {
	id, err := dnTypes.NewIDFromString(argValue)
	if err != nil {
		return dnTypes.ID{}, fmt.Errorf("%s %s %q: %v", argName, paramType, argValue, err)
	}

	return id, nil
}

func AddPaginationCmdFlags(cmd *cobra.Command) {
	cmd.Flags().String(flags.FlagPage, "1", "pagination page of destroys to to query for (first page: 1)")
	cmd.Flags().String(flags.FlagLimit, "100", "pagination limit of destroys to query for")
}
