package helpers

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	cliCtx "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govCli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

type ParamType string

const (
	ParamTypeCliFlag     ParamType = "flag"
	ParamTypeCliArg      ParamType = "argument"
	ParamTypeRestQuery   ParamType = "query param"
	ParamTypeRestPath    ParamType = "path param"
	ParamTypeRestRequest ParamType = "request param"
)

// GetTxCmdCtx context from CLI Tx commands.
func GetTxCmdCtx(cdc *codec.Codec, cmdInputBuf io.Reader) (cliCtx cliCtx.CLIContext, txBuilder authTypes.TxBuilder) {
	inBuf := bufio.NewReader(cmdInputBuf)
	cliCtx = context.NewCLIContextWithInput(inBuf).WithCodec(cdc)
	txBuilder = authTypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

	return
}

// ParseFromFlag parses --from flag.
func ParseFromFlag(cliCtx cliCtx.CLIContext) (sdk.AccAddress, error) {
	accGetter := authTypes.NewAccountRetriever(cliCtx)

	if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
		return sdk.AccAddress{}, fmt.Errorf("%s %s: %w", flags.FlagFrom, ParamTypeCliFlag, err)
	}

	return cliCtx.FromAddress, nil
}

// ParseDepositFlag parses --deposit flag.
func ParseDepositFlag(flags *pflag.FlagSet) (sdk.Coins, error) {
	depositStr, err := flags.GetString(govCli.FlagDeposit)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("%s %s: %w", govCli.FlagDeposit, ParamTypeCliFlag, err)
	}
	deposit, err := sdk.ParseCoins(depositStr)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("%s %s %q: parsing Coins: %w", govCli.FlagDeposit, ParamTypeCliFlag, depositStr, err)
	}

	return deposit, nil
}

// ParsePaginationParams parses --page, --limit flags.
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

// ParseSdkIntParam parses sdk.Int param.
func ParseSdkIntParam(argName, argValue string, paramType ParamType) (sdk.Int, error) {
	v, ok := sdk.NewIntFromString(argValue)
	if !ok {
		return sdk.Int{}, fmt.Errorf("%s %s %q: parsing Int: failed", argName, paramType, argValue)
	}

	return v, nil
}

// ParseSdkIntParam parses sdk.Uint param.
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

// ParseUint8Param parses uint8 param.
func ParseUint8Param(argName, argValue string, paramType ParamType) (uint8, error) {
	v, err := strconv.ParseUint(argValue, 10, 8)
	if err != nil {
		return uint8(0), fmt.Errorf("%s %s %q: uint8 parsing: %w", argName, paramType, argValue, err)
	}

	return uint8(v), nil
}

// ParseUint64Param parses uint64 param.
func ParseUint64Param(argName, argValue string, paramType ParamType) (uint64, error) {
	v, err := strconv.ParseUint(argValue, 10, 64)
	if err != nil {
		return uint64(0), fmt.Errorf("%s %s %q: uint64 parsing: %w", argName, paramType, argValue, err)
	}

	return v, nil
}

// ParseSdkAddressParam parses sdk.AccAddress param.
func ParseSdkAddressParam(argName, argValue string, paramType ParamType) (sdk.AccAddress, error) {
	if v, err := sdk.AccAddressFromBech32(argValue); err == nil {
		return v, nil
	}

	argValueNorm := strings.TrimPrefix(argValue, "0x")
	if v, err := sdk.AccAddressFromHex(argValueNorm); err == nil {
		return v, nil
	}

	return sdk.AccAddress{}, fmt.Errorf("%s %s %q: parsing Bech32 / HEX account address: failed", argName, paramType, argValue)
}

// ParseEthereumAddressParam parses and validates Ethereum address param.
func ParseEthereumAddressParam(argName, argValue string, paramType ParamType) (string, error) {
	if !IsEthereumAddress(argValue) {
		return "", fmt.Errorf("%s %s %q: ethereum address validation failed", argName, paramType, argValue)
	}

	return argValue, nil
}

// ParseDnIDParam parses dnTypes.ID param.
func ParseDnIDParam(argName, argValue string, paramType ParamType) (dnTypes.ID, error) {
	id, err := dnTypes.NewIDFromString(argValue)
	if err != nil {
		return dnTypes.ID{}, fmt.Errorf("%s %s %q: %v", argName, paramType, argValue, err)
	}

	return id, nil
}

// ValidateDenomParam validates currency denomination symbol.
func ValidateDenomParam(argName, argValue string, paramType ParamType) error {
	if err := dnTypes.DenomFilter(argValue); err != nil {
		return fmt.Errorf("%s %s %q: %v", argName, paramType, argValue, err)
	}

	return nil
}

// ParseHexStringParam parses HEX string param.
func ParseHexStringParam(argName, argValue string, paramType ParamType) (string, error) {
	argValueNorm := strings.TrimPrefix(argValue, "0x")
	if _, err := hex.DecodeString(argValueNorm); err != nil {
		return "", fmt.Errorf("%s %s %q: %v", argName, paramType, argValue, err)
	}

	return argValueNorm, nil
}

// ParseAssetCodeParam parses assetCode and validates it.
func ParseAssetCodeParam(argName, argValue string, paramType ParamType) (dnTypes.AssetCode, error) {
	assetCode := dnTypes.AssetCode(strings.ToLower(argValue))
	if err := assetCode.Validate(); err != nil {
		return "", fmt.Errorf("%s %s %q: %v", argName, paramType, argValue, err)
	}

	return assetCode, nil
}

// ParseCoinParam parses sdk.Coin param and validates it.
func ParseCoinParam(argName, argValue string, paramType ParamType) (retCoin sdk.Coin, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("%s %s %q: parsing coin failed", argName, paramType, argValue)
		}
	}()

	coin, err := sdk.ParseCoin(argValue)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("%s %s %q: parsing coin: %v", argName, paramType, argValue, err)
	}

	if err := dnTypes.DenomFilter(coin.Denom); err != nil {
		return sdk.Coin{}, fmt.Errorf("%s %s %q: validating denom: %v", argName, paramType, argValue, err)
	}

	if coin.Amount.LT(sdk.ZeroInt()) {
		return sdk.Coin{}, fmt.Errorf("%s %s %q: amount is LT zero", argName, paramType, argValue)
	}

	return coin, nil
}

// BuildError builds an error in unified error style.
func BuildError(argName, argValue string, paramType ParamType, errMsg string) error {
	return fmt.Errorf("%s %s %q: %s", argName, paramType, argValue, errMsg)
}

// AddPaginationCmdFlags adds --page --limit flags to Cobra command.
func AddPaginationCmdFlags(cmd *cobra.Command) {
	cmd.Flags().String(flags.FlagPage, "1", "pagination page of objects list to to query for (first page: 1)")
	cmd.Flags().String(flags.FlagLimit, "100", "pagination limit of objects list to query for")
}

// BuildCmdHelp add long description to Cobra command using short description and provided strings.
func BuildCmdHelp(cmd *cobra.Command, argDescriptions []string) {
	args := strings.Split(cmd.Use, " ")
	args = args[1:]

	if len(argDescriptions) != len(args) {
		panic(fmt.Errorf("building Help for cmd %q, argDescriptions len mismatch %d / %d: ", cmd.Use, len(argDescriptions), len(args)))
	}

	helpBuilder := strings.Builder{}
	helpBuilder.WriteString(fmt.Sprintf("%s:\n", cmd.Short))
	for argIdx, arg := range args {
		argDesc := argDescriptions[argIdx]
		helpBuilder.WriteString(fmt.Sprintf("  %s - %s\n", arg, argDesc))
	}

	cmd.Long = helpBuilder.String()
}

// PaginateSlice returns slice start/end indices for slice checking int limits.
// Should be used for queries with pagination, where slice objects index doesn't exists.
func PaginateSlice(sliceLen int, page, limit sdk.Uint) (start, end uint64, retErr error) {
	if sliceLen < 0 {
		retErr = fmt.Errorf("sliceLen: LT zero")
		return
	}
	if page.IsZero() {
		retErr = fmt.Errorf("page: is zero")
		return
	}
	if limit.IsZero() {
		retErr = fmt.Errorf("limit: is zero")
		return
	}
	if sliceLen == 0 {
		return
	}

	start = (page.Uint64() - 1) * limit.Uint64()
	end = limit.Uint64() + start

	if start >= uint64(sliceLen) {
		start, end = 0, 0
		return
	}

	if end >= uint64(sliceLen) {
		end = uint64(sliceLen)
	}

	return
}
