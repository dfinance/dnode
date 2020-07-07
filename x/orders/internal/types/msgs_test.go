// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

func TestOrders_PostOrderMsg_Valid(t *testing.T) {
	ownerAddr := sdk.AccAddress("wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h")

	msg := NewMsgPost(ownerAddr, dnTypes.AssetCode("btc_dfi"), Bid, sdk.OneUint(), sdk.OneUint(), 60)
	require.NoError(t, msg.ValidateBasic())
}

func TestOrders_PostOrderMsg_Invalid(t *testing.T) {
	ownerAddr := sdk.AccAddress("wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h")
	assetCode := dnTypes.AssetCode("btc_dfi")
	direction := Bid
	price := sdk.OneUint()
	quantity := sdk.OneUint()
	ttl := uint64(60)

	// owner
	require.Error(t, NewMsgPost(sdk.AccAddress{}, assetCode, direction, price, quantity, ttl).ValidateBasic())

	// assetCode
	require.Error(t, NewMsgPost(ownerAddr, dnTypes.AssetCode(""), direction, price, quantity, ttl).ValidateBasic())

	// direction
	require.Error(t, NewMsgPost(ownerAddr, assetCode, Direction(""), price, quantity, ttl).ValidateBasic())

	// price
	require.Error(t, NewMsgPost(ownerAddr, assetCode, direction, sdk.ZeroUint(), quantity, ttl).ValidateBasic())

	// quantity
	require.Error(t, NewMsgPost(ownerAddr, assetCode, direction, price, sdk.ZeroUint(), ttl).ValidateBasic())

	// ttl
	require.Error(t, NewMsgPost(ownerAddr, assetCode, direction, price, quantity, 0).ValidateBasic())
}

func TestOrders_RevokeOrderMsg_Valid(t *testing.T) {
	ownerAddr := sdk.AccAddress("wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h")

	msg := NewMsgRevokeOrder(ownerAddr, dnTypes.NewIDFromUint64(0))
	require.NoError(t, msg.ValidateBasic())
}

func TestOrders_RevokeOrderMsg_Invalid(t *testing.T) {
	ownerAddr := sdk.AccAddress("wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h")
	orderID := dnTypes.NewIDFromUint64(0)

	// owner
	require.Error(t, NewMsgRevokeOrder(sdk.AccAddress{}, orderID).ValidateBasic())

	// orderID
	require.Error(t, NewMsgRevokeOrder(ownerAddr, dnTypes.ID{}).ValidateBasic())
}
