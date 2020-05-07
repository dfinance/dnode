package types

import sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrInternal = sdkErrors.Register(ModuleName, 100, "internal")
	// Msg.MarketID is invalid.
	ErrWrongMarketID = sdkErrors.Register(ModuleName, 101, "wrong marketID")
	// Msg.Owner is empty.
	ErrWrongOwner = sdkErrors.Register(ModuleName, 102, "wrong owner")
	// Msg.Price is empty.
	ErrWrongPrice = sdkErrors.Register(ModuleName, 103, "wrong price, should be greater that 0")
	// Msg.Quantity is empty.
	ErrWrongQuantity = sdkErrors.Register(ModuleName, 104, "wrong quantity, should be greater that 0")
	// Msg.Ttl == 0.
	ErrWrongTtl = sdkErrors.Register(ModuleName, 105, "wrong TTL [sec], should be greater that 0")
	// Msg.Direction enumb is invalid.
	ErrWrongDirection = sdkErrors.Register(ModuleName, 106, "wrong direction")
	// Order not exists.
	ErrWrongOrderID = sdkErrors.Register(ModuleName, 107, "wrong orderID")
)
