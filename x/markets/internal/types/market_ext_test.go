// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	ccsTypes "github.com/dfinance/dnode/x/cc_storage"
)

type BaseToQuoteQuantityInput struct {
	Market    MarketExtended
	InBaseP   sdk.Uint
	InBaseQ   sdk.Uint
	OutQuoteQ sdk.Uint
	OutErr    bool
}

func checkBaseToQuoteQuantityInputs(t *testing.T, inputs []BaseToQuoteQuantityInput) {
	for i, input := range inputs {
		t.Logf("[%d] BaseDecimals:  %d", i, input.Market.BaseCurrency.Decimals)
		t.Logf("[%d] QuoteDecimals: %d", i, input.Market.QuoteCurrency.Decimals)
		t.Logf("[%d] InBaseP:   %s -> %s", i, input.InBaseP, input.Market.QuoteCurrency.UintToDec(input.InBaseP))
		t.Logf("[%d] InBaseQ:   %s -> %s", i, input.InBaseQ, input.Market.BaseCurrency.UintToDec(input.InBaseQ))
		t.Logf("[%d] OutQuoteQ: %s -> %s", i, input.OutQuoteQ, input.Market.QuoteCurrency.UintToDec(input.OutQuoteQ))

		quoteQ, err := input.Market.BaseToQuoteQuantity(input.InBaseP, input.InBaseQ)

		if input.OutErr {
			require.Error(t, err, "[%d]: error is expected", i)
		} else {
			require.NoError(t, err, "[%d]: error is not expected", i)
			require.True(t, input.OutQuoteQ.Equal(quoteQ), "[%d]: got / expected: %s / %s", i, input.OutQuoteQ, quoteQ)
		}
	}
}

func TestMarkets_BaseToQuoteQuantity(t *testing.T) {
	t.Parallel()

	marketNoDecimals := MarketExtended{
		BaseCurrency:  ccsTypes.Currency{Decimals: 0},
		QuoteCurrency: ccsTypes.Currency{Decimals: 0},
	}

	marketBase2Quote2 := MarketExtended{
		BaseCurrency:  ccsTypes.Currency{Decimals: 2},
		QuoteCurrency: ccsTypes.Currency{Decimals: 2},
	}

	marketBase2Quote3 := MarketExtended{
		BaseCurrency:  ccsTypes.Currency{Decimals: 2},
		QuoteCurrency: ccsTypes.Currency{Decimals: 3},
	}

	inputs := []BaseToQuoteQuantityInput{
		{
			Market:    marketNoDecimals,
			InBaseP:   sdk.NewUint(1),
			InBaseQ:   sdk.NewUint(1000),
			OutQuoteQ: sdk.NewUint(1000),
			OutErr:    false,
		},
		{
			Market:    marketBase2Quote2,
			InBaseP:   sdk.NewUint(100),
			InBaseQ:   sdk.NewUint(100000),
			OutQuoteQ: sdk.NewUint(100000),
			OutErr:    false,
		},
		{
			Market:    marketBase2Quote2,
			InBaseP:   sdk.NewUint(50),
			InBaseQ:   sdk.NewUint(100000),
			OutQuoteQ: sdk.NewUint(50000),
			OutErr:    false,
		},
		{
			Market:    marketBase2Quote2,
			InBaseP:   sdk.NewUint(50),
			InBaseQ:   sdk.NewUint(1000),
			OutQuoteQ: sdk.NewUint(500),
			OutErr:    false,
		},
		{
			Market:    marketBase2Quote3,
			InBaseP:   sdk.NewUint(5000),
			InBaseQ:   sdk.NewUint(100),
			OutQuoteQ: sdk.NewUint(5000),
			OutErr:    false,
		},
		{
			Market:    marketBase2Quote3,
			InBaseP:   sdk.NewUint(5000),
			InBaseQ:   sdk.NewUint(10),
			OutQuoteQ: sdk.NewUint(500),
			OutErr:    false,
		},
		{
			Market:    marketBase2Quote3,
			InBaseP:   sdk.NewUint(5),
			InBaseQ:   sdk.NewUint(1),
			OutQuoteQ: sdk.ZeroUint(),
			OutErr:    true,
		},
	}

	checkBaseToQuoteQuantityInputs(t, inputs)
}
