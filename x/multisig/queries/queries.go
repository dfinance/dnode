package queries

import (
	"wings-blockchain/x/multisig/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

const (
	QueryLastId = "lastId"
)

func NewQuerier(keeper keeper.Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryLastId:
			return queryLastId(ctx, keeper)

		default:
			return nil, sdk.ErrUnknownRequest("unknown query")
		}
	}
}

func queryLastId(ctx sdk.Context, keeper keeper.Keeper) ([]byte, sdk.Error) {
	lastIdRes := QueryLastIdRes{}

	lastIdRes.LastId = keeper.GetLastId(ctx)
	bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), lastIdRes)

	if err != nil {
		panic(err)
	}

	return bz, nil
}