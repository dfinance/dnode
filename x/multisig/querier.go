// Querier for multisig module.
package multisig

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/multisig/types"
)

// Supported queries.
const (
	QueryGetCalls  = "calls"
	QueryGetLastId = "lastId"
	QueryGetCall   = "call"
	QueryGetUnique = "unique"
)

// Creating new Querier.
func NewQuerier(msKeeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryGetCalls:
			return queryGetCalls(msKeeper, ctx)

		case QueryGetLastId:
			return queryGetLastId(msKeeper, ctx)

		case QueryGetCall:
			return queryGetCall(msKeeper, ctx, req)

		case QueryGetUnique:
			return queryGetUnique(msKeeper, ctx, req)

		default:
			return nil, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "unknown query")
		}
	}
}

// Process request to get last id.
func queryGetLastId(msKeeper Keeper, ctx sdk.Context) ([]byte, error) {
	resp := types.LastIdRes{LastId: msKeeper.GetLastId(ctx)}

	bz, err := codec.MarshalJSONIndent(msKeeper.cdc, resp)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "could not marshal result to JSON: %v", err)
	}

	return bz, nil
}

// Process request to get calls.
func queryGetCalls(msKeeper Keeper, ctx sdk.Context) ([]byte, error) {
	calls := make(types.CallsResp, 0)

	start := ctx.BlockHeight() - msKeeper.GetIntervalToExecute(ctx)
	if start < 0 {
		start = 0
	}

	activeIterator := msKeeper.GetQueueIteratorStartEnd(ctx, start, ctx.BlockHeight())
	defer activeIterator.Close()

	for ; activeIterator.Valid(); activeIterator.Next() {
		bs := activeIterator.Value()

		var callId uint64
		if err := ModuleCdc.UnmarshalBinaryLengthPrefixed(bs, &callId); err != nil {
			return nil, sdkErrors.Wrapf(types.ErrInternal, "could not marshal call id: %v", err)
		}

		var callResp types.CallResp
		call, callErr := msKeeper.GetCall(ctx, callId)
		votes, votesErr := msKeeper.GetVotes(ctx, callId)

		if callErr != nil || votesErr != nil {
			return nil, sdkErrors.Wrapf(types.ErrInternal, "could not extract votes for call by id: %v, %v", callErr, votesErr)
		}

		callResp.Call = call
		callResp.Votes = votes

		calls = append(calls, callResp)
	}

	bz, err := codec.MarshalJSONIndent(msKeeper.cdc, calls)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "could not marshal result to JSON: %v", err)
	}

	return bz, nil
}

// Process request to get call.
func queryGetCall(msKeeper Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, error) {
	var params types.CallReq

	if err := ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	call, err := msKeeper.GetCall(ctx, params.CallId)
	if err != nil {
		return nil, err
	}

	votes, err := msKeeper.GetVotes(ctx, params.CallId)
	if err != nil {
		return nil, err
	}

	callResp := types.CallResp{
		Call:  call,
		Votes: votes,
	}

	bz, errMarshal := codec.MarshalJSONIndent(msKeeper.cdc, callResp)
	if errMarshal != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "could not marshal result to JSON: %v", errMarshal)
	}

	return bz, nil
}

// Process query to get call by unique id.
func queryGetUnique(msKeeper Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, error) {
	var params types.UniqueReq
	if err := ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	id, err := msKeeper.GetCallIDByUnique(ctx, params.UniqueId)
	if err != nil {
		return nil, err
	}

	call, err := msKeeper.GetCall(ctx, id)
	if err != nil {
		return nil, err
	}

	votes, err := msKeeper.GetVotes(ctx, id)
	if err != nil {
		return nil, err
	}

	callResp := types.CallResp{
		Call:  call,
		Votes: votes,
	}

	bz, errMarshal := codec.MarshalJSONIndent(msKeeper.cdc, callResp)
	if errMarshal != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "could not marshal result to JSON: %v", errMarshal)
	}

	return bz, nil
}
