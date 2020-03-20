//
package vm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

const (
	// Queries types for querier.
	QueryValue = "value" // Get value by access path.
)

// Create new querier.
func NewQuerier(vmKeeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryValue:
			return queryGetValue(ctx, vmKeeper, req)

		default:
			return nil, sdk.ErrUnknownRequest("unknown query")
		}
	}
}

// Processing query to get value from DS.
func queryGetValue(ctx sdk.Context, vmKeeper Keeper, req abci.RequestQuery) ([]byte, sdk.Error) {
	var queryAccessPath types.QueryAccessPath

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &queryAccessPath); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse query access path: %s", err))
	}

	return vmKeeper.GetValue(ctx, &vm_grpc.VMAccessPath{
		Address: queryAccessPath.Address,
		Path:    queryAccessPath.Path,
	}), nil
}
