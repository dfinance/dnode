package queries

import (
	"wings-blockchain/x/multisig/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"strconv"
	"wings-blockchain/x/multisig/types"
)

const (
	QueryLastId    = "lastId"
	QueryGetCall   = "call"
	QueryGetCalls  = "calls"
	QueryGetUnique = "unique"
)

// Querier for multisig module
func NewQuerier(keeper keeper.Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryLastId:
			return queryLastId(keeper, ctx)

		case QueryGetCall:
			return queryGetCall(keeper, ctx, path[1:])

		case QueryGetCalls:
			return queryGetCalls(keeper, ctx)

        case QueryGetUnique:
            return queryGetCallByUnique(keeper, ctx, path[1:])

		default:
			return nil, sdk.ErrUnknownRequest("unknown query")
		}
	}
}

// Query handler to get last call id
func queryLastId(keeper keeper.Keeper, ctx sdk.Context) ([]byte, sdk.Error) {
	lastIdRes := QueryLastIdRes{}

	lastIdRes.LastId = keeper.GetLastId(ctx)
	bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), lastIdRes)

	if err != nil {
		panic(err)
	}

	return bz, nil
}

// Query handler to get call by id
func queryGetCalls(keeper keeper.Keeper, ctx sdk.Context) ([]byte, sdk.Error) {
	calls := make(QueryCallsResp, 0)

    start := ctx.BlockHeight() - types.IntervalToExecute

    if start < 0 {
        start = 0
    }

    activeIterator := keeper.GetQueueIteratorStartEnd(ctx, start, ctx.BlockHeight())
    defer activeIterator.Close()

    for ; activeIterator.Valid(); activeIterator.Next() {
        bs := activeIterator.Value()

        var callId uint64
        keeper.GetCDC().MustUnmarshalBinaryLengthPrefixed(bs, &callId)
        call, err := makeCallResp(keeper, ctx, callId)

        if err != nil {
            return []byte{}, err
        }

        calls = append(calls, call)
    }

	bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), calls)

	if err != nil {
		panic(err)
	}

	return bz, nil
}

// Query handler to get call by id
func queryGetCallByUnique(keeper keeper.Keeper, ctx sdk.Context, params []string) ([]byte, sdk.Error) {
    id, err := keeper.GetCallIDByUnique(ctx, params[0])

    if err != nil {
        return nil, err
    }

    callResp, err2 := makeCallResp(keeper, ctx, id)

    if err2 !=  nil {
        return nil, err2
    }

    bz, err3 := codec.MarshalJSONIndent(keeper.GetCDC(), callResp)

    if err3 != nil {
        panic(err3)
    }

    return bz, nil
}

// Query handler to get call by id
func queryGetCall(keeper keeper.Keeper, ctx sdk.Context, params []string) ([]byte, sdk.Error) {
	callIdParam := params[0]
	callId, err := strconv.ParseUint(callIdParam, 10, 64)

	if err != nil {
		return nil, types.ErrCantParseCallId(callIdParam)
	}

	callResp, err2 := makeCallResp(keeper, ctx, callId)

	if err2 !=  nil {
		return nil, err2
	}

	bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), callResp)

	if err != nil {
		panic(err)
	}

	return bz, nil
}

// Make a call response from call
func makeCallResp(keeper keeper.Keeper, ctx sdk.Context, callId uint64) (QueryCallResp, sdk.Error) {
	callRes := QueryCallResp{}

	call, err := keeper.GetCall(ctx, callId)

	if err != nil {
		return QueryCallResp{}, err
	}

	callRes.Call = call

	votes, err := keeper.GetVotes(ctx, callId)

	if err != nil {
		return QueryCallResp{}, err
	}

	callRes.Votes = votes

	return callRes, nil
}
