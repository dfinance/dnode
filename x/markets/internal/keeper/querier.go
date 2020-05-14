package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/markets/internal/types"
)

const (
	QueryList = "list"
)

// NewQuerier return keeper querier.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case QueryList:
			return queryList(ctx, k)
		default:
			return nil, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "unknown markets query endpoint")
		}
	}
}

// queryList handles list query which return all market objects.
func queryList(ctx sdk.Context, k Keeper) ([]byte, error) {
	markets := types.Markets{}
	markets = append(markets, k.GetParams(ctx).Markets...)

	res, err := codec.MarshalJSONIndent(k.cdc, markets)
	if err != nil {
		return nil, fmt.Errorf("markets marshal: %w", err)
	}

	k.GetLogger(ctx).Debug(fmt.Sprintf("Markets table:\n%s", markets.String()))

	return res, nil
}
