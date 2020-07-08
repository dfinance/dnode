package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// NewHandler handles all oracle type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case types.MsgPostPrice:
			return handleMsgPostPrice(ctx, k, msg)
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

// handleMsgPostPrice handles prices posted by oracles.
func handleMsgPostPrice(ctx sdk.Context, k Keeper, msg types.MsgPostPrice) (*sdk.Result, error) {
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

// handleMsgAddOracle handles AddOracle message.
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

// handleMsgSetOracles handles SetOracles message.
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

// handleMsgSetAsset handles SetAsset message.
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

//  handleMsgAddAsset handles AddUser message.
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
