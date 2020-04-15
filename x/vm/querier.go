//
package vm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

const (
	// Queries types for querier.
	QueryValue = "value" // Get value by access path.
)

// Create new querier.
func NewQuerier(vmKeeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryValue:
			return queryGetValue(ctx, vmKeeper, req)

		default:
			return nil, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "unknown query")
		}
	}
}

// Processing query to get value from DS.
func queryGetValue(ctx sdk.Context, vmKeeper Keeper, req abci.RequestQuery) ([]byte, error) {
	var queryAccessPath types.QueryAccessPath

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &queryAccessPath); err != nil {
		return nil, sdkErrors.Wrap(types.ErrInternal, "unknown query")
	}

	return vmKeeper.GetValue(ctx, &vm_grpc.VMAccessPath{
		Address: queryAccessPath.Address,
		Path:    queryAccessPath.Path,
	}), nil
}
