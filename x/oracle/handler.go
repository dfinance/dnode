package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// NewHandler handles all oracle type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
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
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized oracle message type: %T", msg)
		}
	}
}

// price feed questions:
// do proposers need to post the round in the message? If not, how do we determine the round?

// HandleMsgPostPrice handles prices posted by oracles
func HandleMsgPostPrice(ctx sdk.Context, k Keeper, msg types.MsgPostPrice) (*sdk.Result, error) {
	// TODO cleanup message validation and errors
	if err := k.ValidatePostPrice(ctx, msg); err != nil {
		return nil, err
	}

	if _, err := k.GetOracle(ctx, msg.AssetCode, msg.From); err != nil {
		return nil, sdkErrors.Wrap(types.ErrInvalidOracle, msg.From.String())
	}

	if _, err := k.SetPrice(ctx, msg.From, msg.AssetCode, msg.Price, msg.ReceivedAt); err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgAddOracle(ctx sdk.Context, k Keeper, msg types.MsgAddOracle) (*sdk.Result, error) {
	// TODO cleanup message validation and errors
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if _, err := k.GetOracle(ctx, msg.Denom, msg.Oracle); err == nil {
		return nil, sdkErrors.Wrap(types.ErrInvalidOracle, msg.Oracle.String())
	}

	if err := k.AddOracle(ctx, msg.Nominee.String(), msg.Denom, msg.Oracle); err != nil {
		return nil, sdkErrors.Wrap(types.ErrInternal, err.Error())
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgSetOracles(ctx sdk.Context, k Keeper, msg types.MsgSetOracles) (*sdk.Result, error) {
	// TODO cleanup message validation and errors
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if _, found := k.GetAsset(ctx, msg.Denom); !found {
		return nil, sdkErrors.Wrap(types.ErrInvalidAsset, msg.Denom)
	}

	if err := k.SetOracles(ctx, msg.Nominee.String(), msg.Denom, msg.Oracles); err != nil {
		return nil, sdkErrors.Wrap(types.ErrInternal, err.Error())
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgSetAsset(ctx sdk.Context, k Keeper, msg types.MsgSetAsset) (*sdk.Result, error) {
	// TODO cleanup message validation and errors
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if _, found := k.GetAsset(ctx, msg.Denom); !found {
		return nil, sdkErrors.Wrap(types.ErrInvalidAsset, msg.Denom)
	}

	if err := k.SetAsset(ctx, msg.Nominee.String(), msg.Denom, msg.Asset); err != nil {
		return nil, sdkErrors.Wrap(types.ErrInternal, err.Error())
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgAddAsset(ctx sdk.Context, k Keeper, msg types.MsgAddAsset) (*sdk.Result, error) {
	// TODO cleanup message validation and errors
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if _, found := k.GetAsset(ctx, msg.Denom); found {
		return nil, sdkErrors.Wrap(types.ErrExistingAsset, msg.Denom)
	}

	if err := k.AddAsset(ctx, msg.Nominee.String(), msg.Denom, msg.Asset); err != nil {
		return nil, sdkErrors.Wrap(types.ErrInternal, err.Error())
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
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
