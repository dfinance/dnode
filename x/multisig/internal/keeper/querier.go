package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// NewQuerier return keeper querier.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryCalls:
			return queryGetCalls(k, ctx)
		case types.QueryCall:
			return queryGetCall(k, ctx, req)
		case types.QueryCallByUnique:
			return queryGetUnique(k, ctx, req)
		case types.QueryLastId:
			return queryGetLastID(k, ctx)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unsupported query endpoint %q for module %q", path[0], types.ModuleName)
		}
	}
}

// queryGetCalls handles getCalls query which returns call objects.
func queryGetCalls(k Keeper, ctx sdk.Context) ([]byte, error) {
	resps := types.CallsResp{}

	// define range start
	start := ctx.BlockHeight() - k.GetIntervalToExecute(ctx)
	if start < 0 {
		start = 0
	}

	iterator := k.GetQueueIteratorStartEnd(ctx, start, ctx.BlockHeight())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()

		var callID dnTypes.ID
		if err := types.ModuleCdc.UnmarshalBinaryLengthPrefixed(bz, &callID); err != nil {
			return nil, sdkErrors.Wrapf(types.ErrInternal, "callID unmarshal: %v", err)
		}

		call, err := k.GetCall(ctx, callID)
		if err != nil {
			return nil, err
		}

		votes, err := k.GetVotes(ctx, callID)
		if err != nil {
			return nil, err
		}

		resps = append(resps, types.CallResp{
			Call:  call,
			Votes: votes,
		})
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, resps)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "[]callsRes marshal: %v", err)
	}

	return bz, nil
}

// queryGetCall handles getCall query which returns call object.
func queryGetCall(k Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, error) {
	var params types.CallReq
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	call, err := k.GetCall(ctx, params.CallID)
	if err != nil {
		return nil, err
	}

	votes, err := k.GetVotes(ctx, params.CallID)
	if err != nil {
		return nil, err
	}

	resp := types.CallResp{
		Call:  call,
		Votes: votes,
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, resp)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "callResp marshal: %v", err)
	}

	return bz, nil
}

// queryGetUnique handles QueryCallByUnique query which returns call object by its uniqueID.
// Process query to get call by unique id.
func queryGetUnique(k Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, error) {
	var params types.CallByUniqueIdReq
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	id, err := k.GetCallIDByUniqueID(ctx, params.UniqueID)
	if err != nil {
		return nil, err
	}

	call, err := k.GetCall(ctx, id)
	if err != nil {
		return nil, err
	}

	votes, err := k.GetVotes(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := types.CallResp{
		Call:  call,
		Votes: votes,
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, resp)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "CallResp marshal: %v", err)
	}

	return bz, nil
}

// queryGetLastID handles getGetLastID query which returns last callID.
func queryGetLastID(k Keeper, ctx sdk.Context) ([]byte, error) {
	resp := types.LastCallIdResp{LastID: k.GetLastCallID(ctx)}

	bz, err := codec.MarshalJSONIndent(k.cdc, resp)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "lastCallIdResp marshal: %v", err)
	}

	return bz, nil
}
