package keeper

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/dfinance/glav"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// NewQuerier return keeper querier.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryValue:
			return queryGetValue(ctx, k, req)
		case types.QueryLcsView:
			return queryLcsView(ctx, k, req)
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

// queryLcsView handles lcsView query which builds VM path, reads writeSet and tries to represent raw LCS data based on request structure.
func queryLcsView(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var request types.LcsViewReq
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &request); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	// Build VM path using GLAV
	var addrLibra [20]byte
	copy(addrLibra[:], common_vm.Bech32ToLibra(request.Address)[:20])
	resPath := glav.NewStructTag(addrLibra, request.ModuleName, request.StructName, nil).AccessVector()

	// Get raw writeSet data
	resData := k.GetValueWithMiddlewares(ctx, &vm_grpc.VMAccessPath{Address: request.Address, Path: resPath})
	if resData == nil {
		return nil, sdkErrors.Wrapf(types.ErrNotFound, "data at accessPath 0x%s::%s: not found", hex.EncodeToString(addrLibra[:]), hex.EncodeToString(resPath))
	}

	// Get LCS view
	resp, err := StringifyLCSData(request.ViewRequest, resData)
	if err != nil {
		return nil, sdkErrors.Wrap(types.ErrInternal, err.Error())
	}

	return []byte(resp), nil
}
