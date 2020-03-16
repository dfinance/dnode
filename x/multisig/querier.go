// Querier for multisig module.
package multisig

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
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
			return nil, sdk.ErrUnknownRequest("unknown query")
		}
	}
}

// Process request to get last id.
func queryGetLastId(msKeeper Keeper, ctx sdk.Context) ([]byte, sdk.Error) {
	resp := types.LastIdRes{LastId: msKeeper.GetLastId(ctx)}

	bz, err := codec.MarshalJSONIndent(msKeeper.cdc, resp)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// Process request to get calls.
func queryGetCalls(msKeeper Keeper, ctx sdk.Context) ([]byte, sdk.Error) {
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
		err := ModuleCdc.UnmarshalBinaryLengthPrefixed(bs, &callId)

		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal call id", err.Error()))
		}

		var callResp types.CallResp
		call, err := msKeeper.GetCall(ctx, callId)
		votes, err := msKeeper.GetVotes(ctx, callId)

		if err != nil {
			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not extract votes for call by id", err.Error()))
		}

		callResp.Call = call
		callResp.Votes = votes

		calls = append(calls, callResp)
	}

	bz, err := codec.MarshalJSONIndent(msKeeper.cdc, calls)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// Process request to get call.
func queryGetCall(msKeeper Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, sdk.Error) {
	var params types.CallReq

	if err := ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
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
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", errMarshal.Error()))
	}

	return bz, nil
}

// Process query to get call by unique id.
func queryGetUnique(msKeeper Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, sdk.Error) {
	var params types.UniqueReq
	if err := ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
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
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", errMarshal.Error()))
	}

	return bz, nil
}
