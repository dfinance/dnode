package keeper

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/helpers"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// GetCurrentPrice fetches the current median price of all oracles for a specific asset.
func (k Keeper) GetCurrentPrice(ctx sdk.Context, assetCode dnTypes.AssetCode) types.CurrentPrice {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.CurrentPricePrefix + assetCode))

	var price types.CurrentPrice
	k.cdc.MustUnmarshalBinaryBare(bz, &price)

	return price
}

// GetRawPrices fetches the set of all prices posted by oracles for an asset and specific blockHeight.
func (k Keeper) GetRawPrices(ctx sdk.Context, assetCode dnTypes.AssetCode, blockHeight int64) []types.PostedPrice {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRawPricesKey(assetCode, blockHeight))

	var prices []types.PostedPrice
	k.cdc.MustUnmarshalBinaryBare(bz, &prices)

	return prices
}

// Check PostPrice's ReceivedAt timestamp (algorithm depends on module params)
func (k Keeper) CheckPriceReceivedAtTimestamp(ctx sdk.Context, receivedAt time.Time) error {
	cfg := k.GetPostPriceParams(ctx)

	if cfg.ReceivedAtDiffInS > 0 {
		thresholdDur := time.Duration(cfg.ReceivedAtDiffInS) * time.Second

		absDuration := func(dur time.Duration) time.Duration {
			if dur < 0 {
				return -dur
			}
			return dur
		}

		blockTime := ctx.BlockTime()
		diffDur := blockTime.Sub(receivedAt)
		if absDuration(diffDur) > thresholdDur {
			return sdkErrors.Wrapf(types.ErrInvalidReceivedAt, "timestamp difference %v should be less than %v", diffDur, thresholdDur)
		}
	}

	return nil
}

// SetPrice updates the posted price for a specific oracle.
func (k Keeper) SetPrice(
	ctx sdk.Context,
	oracle sdk.AccAddress,
	assetCode dnTypes.AssetCode,
	price sdk.Int,
	receivedAt time.Time) (types.PostedPrice, error) {

	// validate price receivedAt timestamp comparing to the current blockHeight timestamp
	if err := k.CheckPriceReceivedAtTimestamp(ctx, receivedAt); err != nil {
		return types.PostedPrice{}, err
	}

	// find raw price for specified oracle
	store := ctx.KVStore(k.storeKey)
	prices := k.GetRawPrices(ctx, assetCode, ctx.BlockHeight())
	var index int
	found := false
	for i := range prices {
		if prices[i].OracleAddress.Equals(oracle) {
			index = i
			found = true
			break
		}
	}

	// set the rawPrice for that particular oracle
	if found {
		prices[index] = types.PostedPrice{
			AssetCode: assetCode, OracleAddress: oracle,
			Price: price, ReceivedAt: receivedAt}
	} else {
		prices = append(prices, types.PostedPrice{
			AssetCode: assetCode, OracleAddress: oracle,
			Price: price, ReceivedAt: receivedAt})
		index = len(prices) - 1
	}

	store.Set(types.GetRawPricesKey(assetCode, ctx.BlockHeight()), k.cdc.MustMarshalBinaryBare(prices))

	return prices[index], nil
}

// SetCurrentPrices updates the price of an asset to the median of all valid oracle inputs and cleans up previous inputs.
func (k Keeper) SetCurrentPrices(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)
	assets := k.GetAssetParams(ctx)

	updatesCnt := 0
	for _, v := range assets {
		assetCode := v.AssetCode
		rawPrices := k.GetRawPrices(ctx, assetCode, ctx.BlockHeight())

		l := len(rawPrices)
		var medianPrice sdk.Int
		var medianReceivedAt time.Time
		// TODO make threshold for acceptance (ie. require 51% of oracles to have posted valid prices
		if l == 0 {
			// Error if there are no valid prices in the raw oracle
			//return types.ErrNoValidPrice(k.codespace)
			medianPrice = sdk.ZeroInt()
		} else if l == 1 {
			// Return immediately if there's only one price
			medianPrice, medianReceivedAt = rawPrices[0].Price, rawPrices[0].ReceivedAt
		} else {
			// sort the prices
			sort.Slice(rawPrices, func(i, j int) bool {
				return rawPrices[i].Price.LT(rawPrices[j].Price)
			})
			// If there's an even number of prices
			if l%2 == 0 {
				// TODO make sure this is safe.
				// Since it's a price and not a balance, division with precision loss is OK.
				price1 := rawPrices[l/2-1].Price
				price2 := rawPrices[l/2].Price
				sum := price1.Add(price2)
				divsor := sdk.NewInt(2)
				medianPrice = sum.Quo(divsor)
				medianReceivedAt = ctx.BlockTime().UTC()
			} else {
				// integer division, so we'll get an integer back, rounded down
				medianPrice, medianReceivedAt = rawPrices[l/2].Price, rawPrices[l/2].ReceivedAt
			}
		}

		// check if there is no rawPrices or medianPrice is invalid
		if medianPrice.IsZero() {
			continue
		}

		// check new price for the asset appeared, no need to update after every block
		oldPrice := k.GetCurrentPrice(ctx, assetCode)
		if oldPrice.AssetCode != "" && oldPrice.Price.Equal(medianPrice) {
			continue
		}

		// set the new price for the asset
		newPrice := types.CurrentPrice{
			AssetCode:  assetCode,
			Price:      medianPrice,
			ReceivedAt: medianReceivedAt,
		}

		store.Set([]byte(types.CurrentPricePrefix+assetCode), k.cdc.MustMarshalBinaryBare(newPrice))

		// save price to vm storage
		accessPath := k.vmKeeper.GetOracleAccessPath(newPrice.AssetCode)
		k.vmKeeper.SetValue(ctx, accessPath, helpers.BigToBytes(newPrice.Price, types.PriceBytesLimit))

		// emit event
		updatesCnt++
		ctx.EventManager().EmitEvent(types.NewPriceEvent(newPrice))
	}

	if updatesCnt > 0 {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		))
	}

	return nil
}

// nolint:errcheck
// ValidatePostPrice makes sure the person posting the price is an oracle.
func (k Keeper) ValidatePostPrice(ctx sdk.Context, msg types.MsgPostPrice) error {
	// TODO implement this

	_, assetFound := k.GetAsset(ctx, msg.AssetCode)
	if !assetFound {
		return sdkErrors.Wrap(types.ErrInvalidAsset, msg.AssetCode.String())
	}
	_, err := k.GetOracle(ctx, msg.AssetCode, msg.From)
	if err != nil {
		return sdkErrors.Wrap(types.ErrInvalidOracle, msg.From.String())
	}

	return nil
}
