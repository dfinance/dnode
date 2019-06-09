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
	QueryLastId   = "lastId"
	QueryGetCall  = "call"
)

// Querier for multisig module
func NewQuerier(keeper keeper.Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryLastId:
			return queryLastId(keeper, ctx)

		case QueryGetCall:
			return queryGetCalls(keeper, ctx, path[1:])

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
func queryGetCalls(keeper keeper.Keeper, ctx sdk.Context, params []string) ([]byte, sdk.Error) {
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

	return callRes, nil
}
