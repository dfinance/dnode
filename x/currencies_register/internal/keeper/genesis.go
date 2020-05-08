package keeper

import (
	"encoding/hex"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

var (
	genesisKey = []byte("genesis")
)

// Init genesis from json.
func (keeper Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) error {
	store := ctx.KVStore(keeper.storeKey)

	var state types.GenesisState
	err := types.ModuleCdc.UnmarshalJSON(data, &state)
	if err != nil {
		return err
	}

	for _, genCurr := range state.Currencies {
		bzPath, err := hex.DecodeString(genCurr.Path)
		if err != nil {
			return err
		}

		err = keeper.AddCurrencyInfo(ctx, genCurr.Denom, genCurr.Decimals, false, types.DefaultOwner, genCurr.TotalSupply, bzPath)
		if err != nil {
			return err
		}
	}

	store.Set(genesisKey, data)

	return nil
}

// Export initialized genesis.
func (keeper Keeper) ExportGenesis(ctx sdk.Context) json.RawMessage {
	store := ctx.KVStore(keeper.storeKey)

	return store.Get(genesisKey)
}
