package types

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/olekukonko/tablewriter"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/markets"
)

// Market order object type.
type Order struct {
	// Order unique ID
	ID dnTypes.ID `json:"id" yaml:"id" example:"0" format:"string representation for big.Uint" swaggertype:"string"`
	// Order owner account address
	Owner sdk.AccAddress `json:"owner" yaml:"owner" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	// Market order belong to
	Market markets.MarketExtended `json:"market" yaml:"market"`
	// Order type (bid/ask)
	Direction Direction `json:"direction" yaml:"direction" swaggertype:"string" example:"bid"`
	// Order target price (in quote asset denom)
	Price sdk.Uint `json:"price" yaml:"price" swaggertype:"string" example:"100"`
	// Order target quantity
	Quantity sdk.Uint `json:"quantity" yaml:"quantity" swaggertype:"string" example:"50"`
	// TimeToLive order auto-cancel period
	Ttl time.Duration `json:"ttl_dur" yaml:"ttl_dur" swaggertype:"integer" example:"60"`
	// Created timestamp
	CreatedAt time.Time `json:"created_at" yaml:"created_at" format:"RFC 3339" example:"2020-03-27T13:45:15.293426Z"`
	// Updated timestamp
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at" format:"RFC 3339" example:"2020-03-27T13:45:15.293426Z"`
}

// Valid checks that Order is valid (used for genesis ops).
func (o Order) Valid() error {
	if err := o.ID.Valid(); err != nil {
		return fmt.Errorf("id: %w", err)
	}
	if o.Owner.Empty() {
		return fmt.Errorf("owner: empty")
	}
	if err := sdk.VerifyAddressFormat(o.Owner); err != nil {
		return fmt.Errorf("owner address format is wrong: %w", err)
	}
	if err := o.Market.Valid(); err != nil {
		return fmt.Errorf("market: %w", err)
	}
	if !o.Direction.IsValid() {
		return fmt.Errorf("direction: invalid")
	}
	if o.Price.IsZero() {
		return fmt.Errorf("price: is zero")
	}
	if o.Quantity.IsZero() {
		return fmt.Errorf("quantity: is zero")
	}
	if o.CreatedAt.After(o.UpdatedAt) {
		return fmt.Errorf("wrong create and update dates: create date later than update date")
	}
	if o.CreatedAt.IsZero() {
		return fmt.Errorf("created_at: is zero")
	}
	if o.CreatedAt.After(time.Now()) {
		return fmt.Errorf("created_at: is future date")
	}
	if o.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at: is zero")
	}
	if o.UpdatedAt.After(time.Now()) {
		return fmt.Errorf("updated_at: is future date")
	}
	return nil
}

// ValidatePriceQuantity compares price and quantity to min currency values.
func (o Order) ValidatePriceQuantity() error {
	minQuotePrice := o.Market.QuoteCurrency.MinDecimal()
	quotePrice := o.Market.QuoteCurrency.UintToDec(o.Price)

	minBaseQuantity := o.Market.BaseCurrency.MinDecimal()
	baseQuantity := o.Market.BaseCurrency.UintToDec(o.Quantity)

	if quotePrice.LT(minQuotePrice) {
		return sdkErrors.Wrapf(ErrWrongPrice, "should be GTE than %s", minQuotePrice.String())
	}
	if baseQuantity.LT(minBaseQuantity) {
		return sdkErrors.Wrapf(ErrWrongQuantity, "should be GTE than %s", minBaseQuantity.String())
	}

	return nil
}

// LockCoin return Coin that should be locked (transferred from account to the module).
// Coin denom and quantity are Marked and Order type specific.
func (o Order) LockCoin() (retCoin sdk.Coin, retErr error) {
	var coinDenom string
	var coinQuantity sdk.Int

	switch o.Direction {
	case Bid:
		quantity, err := o.Market.BaseToQuoteQuantity(o.Price, o.Quantity)
		if err != nil {
			retErr = err
			return
		}
		coinDenom, coinQuantity = o.Market.QuoteDenom(), sdk.NewIntFromBigInt(quantity.BigInt())
	case Ask:
		coinDenom, coinQuantity = o.Market.BaseDenom(), sdk.NewIntFromBigInt(o.Quantity.BigInt())
	default:
		retErr = sdkErrors.Wrap(ErrWrongDirection, o.Direction.String())
		return
	}

	retCoin = sdk.NewCoin(coinDenom, coinQuantity)

	return
}

// Strings returns multi-line text object representation.
func (o Order) String() string {
	b := strings.Builder{}
	b.WriteString("Order:\n")
	b.WriteString(fmt.Sprintf("  ID:        %s\n", o.ID.String()))
	b.WriteString(fmt.Sprintf("  Owner:     %s\n", o.Owner.String()))
	b.WriteString(fmt.Sprintf("  Direction: %s\n", o.Direction.String()))
	b.WriteString(fmt.Sprintf("  Price:     %s\n", o.Price.String()))
	if o.Direction == Bid {
		b.WriteString(fmt.Sprintf("  QQuantity: %s\n", o.Market.QuoteCurrency.UintToDec(o.Quantity).String()))
	} else {
		b.WriteString(fmt.Sprintf("  BQuantity: %s\n", o.Market.BaseCurrency.UintToDec(o.Quantity).String()))
	}
	b.WriteString(fmt.Sprintf("  Ttl:       %s\n", o.Ttl.String()))
	b.WriteString(fmt.Sprintf("  CreatedAt: %s\n", o.CreatedAt.String()))
	b.WriteString(fmt.Sprintf("  UpdatedAt: %s\n", o.UpdatedAt.String()))
	b.WriteString(o.Market.String())

	return b.String()
}

// TableHeaders returns table headers for multi-line text table object representation.
func (o Order) TableHeaders() []string {
	h := []string{
		"O.ID",
		"O.Owner",
		"O.Direction",
		"O.Price",
		"O.QBQuantity",
		"O.TTL",
		"O.CreatedAt",
		"O.UpdatedAt",
	}

	return append(h, o.Market.TableHeaders()...)
}

// TableHeaders returns table rows for multi-line text table object representation.
func (o Order) TableValues() []string {
	v := []string{
		o.ID.String(),
		o.Owner.String(),
		o.Direction.String(),
		o.Price.String(),
	}
	if o.Direction == Bid {
		v = append(v, o.Market.QuoteCurrency.UintToDec(o.Quantity).String())
	} else {
		v = append(v, o.Market.BaseCurrency.UintToDec(o.Quantity).String())
	}
	v = append(v, o.Ttl.String())
	v = append(v, o.CreatedAt.String())
	v = append(v, o.UpdatedAt.String())

	return append(v, o.Market.TableValues()...)
}

// NewOrder creates a new order object.
func NewOrder(
	ctx sdk.Context,
	id dnTypes.ID,
	owner sdk.AccAddress,
	market markets.MarketExtended,
	direction Direction,
	price sdk.Uint,
	quantity sdk.Uint,
	ttlInSec uint64) Order {

	return Order{
		ID:        id,
		Owner:     owner,
		Market:    market,
		Direction: direction,
		Price:     price,
		Quantity:  quantity,
		Ttl:       time.Duration(ttlInSec) * time.Second,
		CreatedAt: ctx.BlockTime(),
		UpdatedAt: ctx.BlockTime(),
	}
}

// Order slice type.
type Orders []Order

// Strings returns multi-line text object representation.
func (l Orders) String() string {
	var buf bytes.Buffer

	t := tablewriter.NewWriter(&buf)
	t.SetHeader(Order{}.TableHeaders())

	for _, o := range l {
		t.Append(o.TableValues())
	}
	t.Render()

	return buf.String()
}
