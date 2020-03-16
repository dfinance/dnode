package oracle

import (
	"fmt"

	"github.com/WingsDao/wings-blockchain/x/oracle/internal/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewHandler handles all oracle type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgPostPrice:
			return HandleMsgPostPrice(ctx, k, msg)
		case types.MsgAddOracle:
			return handleMsgAddOracle(ctx, k, msg)
		case types.MsgSetOracles:
			return handleMsgSetOracles(ctx, k, msg)
		case types.MsgSetAsset:
			return handleMsgSetAsset(ctx, k, msg)
		case types.MsgAddAsset:
			return handleMsgAddAsset(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized oracle message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// price feed questions:
// do proposers need to post the round in the message? If not, how do we determine the round?

// HandleMsgPostPrice handles prices posted by oracles
func HandleMsgPostPrice(
	ctx sdk.Context,
	k Keeper,
	msg types.MsgPostPrice) sdk.Result {

	// TODO cleanup message validation and errors
	if err := k.ValidatePostPrice(ctx, msg); err != nil {
		return err.Result()
	}
	_, er := k.GetOracle(ctx, msg.AssetCode, msg.From)
	if er != nil {
		return types.ErrInvalidOracle(k.Codespace()).Result()
	}
	if _, err := k.SetPrice(ctx, msg.From, msg.AssetCode, msg.Price, msg.ReceivedAt); err != nil {
		return err.Result()
	}

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgAddOracle(
	ctx sdk.Context,
	k Keeper,
	msg types.MsgAddOracle) sdk.Result {

	// TODO cleanup message validation and errors
	if err := msg.ValidateBasic(); err != nil {
		return err.Result()
	}
	if _, err := k.GetOracle(ctx, msg.Denom, msg.Oracle); err == nil {
		return types.ErrInvalidOracle(k.Codespace()).Result()
	}
	if err := k.AddOracle(ctx, msg.Nominee.String(), msg.Denom, msg.Oracle); err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgSetOracles(
	ctx sdk.Context,
	k Keeper,
	msg types.MsgSetOracles) sdk.Result {

	// TODO cleanup message validation and errors
	if err := msg.ValidateBasic(); err != nil {
		return err.Result()
	}
	_, found := k.GetAsset(ctx, msg.Denom)
	if !found {
		return types.ErrInvalidAsset(k.Codespace()).Result()
	}
	er := k.SetOracles(ctx, msg.Nominee.String(), msg.Denom, msg.Oracles)
	if er != nil {
		return sdk.ErrInternal(er.Error()).Result()
	}
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgSetAsset(
	ctx sdk.Context,
	k Keeper,
	msg types.MsgSetAsset) sdk.Result {

	// TODO cleanup message validation and errors
	if err := msg.ValidateBasic(); err != nil {
		return err.Result()
	}
	_, found := k.GetAsset(ctx, msg.Denom)
	if !found {
		return types.ErrInvalidAsset(k.Codespace()).Result()
	}
	er := k.SetAsset(ctx, msg.Nominee.String(), msg.Denom, msg.Asset)
	if er != nil {
		return sdk.ErrInternal(er.Error()).Result()
	}
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgAddAsset(
	ctx sdk.Context,
	k Keeper,
	msg types.MsgAddAsset) sdk.Result {

	// TODO cleanup message validation and errors
	if err := msg.ValidateBasic(); err != nil {
		return err.Result()
	}
	_, found := k.GetAsset(ctx, msg.Denom)
	if found {
		return types.ErrExistingAsset(k.Codespace()).Result()
	}
	er := k.AddAsset(ctx, msg.Nominee.String(), msg.Denom, msg.Asset)
	if er != nil {
		return sdk.ErrInternal(er.Error()).Result()
	}
	return sdk.Result{Events: ctx.EventManager().Events()}
}

// nolint:errcheck
// EndBlocker updates the current oracle
func EndBlocker(ctx sdk.Context, k Keeper) []abci.ValidatorUpdate {
	// TODO val_state_change.go is relevant if we want to rotate the oracle set

	// Running in the end blocker ensures that prices will update at most once per block,
	// which seems preferable to having state storage values change in response to multiple transactions
	// which occur during a block
	//TODO use an iterator and update the prices for all assets in the store
	k.SetCurrentPrices(ctx)
	return []abci.ValidatorUpdate{}
}
