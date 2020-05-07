package types

import (
	"bytes"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/olekukonko/tablewriter"
)

type OrderFill struct {
	Order            Order
	ClearancePrice   sdk.Uint
	QuantityFilled   sdk.Uint
	QuantityUnfilled sdk.Uint
}

func (f OrderFill) FillCoin() (retCoin sdk.Coin, retErr error) {
	coinDenom, coinQuantity := "", sdk.Int{}

	switch f.Order.Direction {
	case Bid:
		coinDenom, coinQuantity = f.Order.Market.BaseAssetDenom, sdk.NewIntFromBigInt(f.QuantityFilled.BigInt())
	case Ask:
		quantity, err := f.Order.Market.BaseToQuoteQuantity(f.ClearancePrice, f.QuantityFilled)
		if err != nil {
			retErr = err
			return
		}
		coinDenom, coinQuantity = f.Order.Market.QuoteAssetDenom, sdk.NewIntFromBigInt(quantity.BigInt())
	default:
		retErr = sdkErrors.Wrap(ErrWrongDirection, f.Order.Direction.String())
		return
	}

	retCoin = sdk.NewCoin(coinDenom, coinQuantity)

	return
}

func (f OrderFill) RefundCoin() (doRefund bool, retCoin *sdk.Coin, retErr error) {
	switch f.Order.Direction {
	case Bid:
		if f.ClearancePrice.LT(f.Order.Price) {
			doRefund = true

			priceDiff := f.Order.Price.Sub(f.ClearancePrice)
			quantity, err := f.Order.Market.BaseToQuoteQuantity(priceDiff, f.QuantityFilled)
			if err == nil {
				coin := sdk.NewCoin(f.Order.Market.QuoteAssetDenom, sdk.NewIntFromBigInt(quantity.BigInt()))
				retCoin = &coin
			}
		}
	case Ask:
	default:
		retErr = sdkErrors.Wrap(ErrWrongDirection, f.Order.Direction.String())
		return
	}

	return
}

func (f OrderFill) String() string {
	b := strings.Builder{}
	b.WriteString("OrderFill:\n")
	b.WriteString(fmt.Sprintf("  ClearancePrice: %s\n", f.ClearancePrice.String()))
	b.WriteString(fmt.Sprintf("  Filled:   %s\n", f.QuantityFilled.String()))
	b.WriteString(fmt.Sprintf("  Unfilled: %s\n", f.QuantityUnfilled.String()))
	b.WriteString(f.Order.String())

	return b.String()
}

func (f OrderFill) TableHeaders() []string {
	h := []string{
		"F.ClearancePrice",
		"F.Filled",
		"F.Unfilled",
	}

	return append(h, f.Order.TableHeaders()...)
}

func (f OrderFill) TableValues() []string {
	v := []string{
		f.ClearancePrice.String(),
		f.QuantityFilled.String(),
		f.QuantityUnfilled.String(),
	}

	return append(v, f.Order.TableValues()...)
}

type OrderFills []OrderFill

func (f OrderFills) String() string {
	var buf bytes.Buffer

	t := tablewriter.NewWriter(&buf)
	t.SetHeader(OrderFill{}.TableHeaders())

	for _, o := range f {
		t.Append(o.TableValues())
	}
	t.Render()

	return string(buf.Bytes())
}
