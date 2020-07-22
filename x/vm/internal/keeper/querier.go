package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

// NewQuerier return keeper querier.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryValue:
			return queryGetValue(ctx, k, req)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unsupported query endpoint %q for module %q", path[0], types.ModuleName)
		}
	}
}

// queryGetValue handles getValue query which return writeSet by VM accessPath.
func queryGetValue(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var queryAccessPath types.ValueReq
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &queryAccessPath); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	return k.GetValueWithMiddlewares(ctx, &vm_grpc.VMAccessPath{
		Address: queryAccessPath.Address,
		Path:    queryAccessPath.Path,
	}), nil
}
