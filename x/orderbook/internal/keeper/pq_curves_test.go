// +build unit

package keeper

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const (
	SDAggTypeBid = 0
	SDAggTypeAsk = 1
)

type SDCurvesInput struct {
	Aggs   []SDAggInputItem
	Output []SDOutputItem
	IsOk   bool
}

type SDAggInputItem struct {
	P uint64
	Q uint64
	T int
}

type SDOutputItem struct {
	P uint64
	S uint64
	D uint64
}

func (input SDCurvesInput) Check(t *testing.T) {
	bidAggs, askAggs := OrderAggregates{}, OrderAggregates{}
	for _, item := range input.Aggs {
		price := sdk.NewUint(item.P)
		quantity := sdk.NewUint(item.Q)
		switch item.T {
		case SDAggTypeBid:
			bidAggs = append(bidAggs, OrderAggregate{Price: price, Quantity: quantity})
		case SDAggTypeAsk:
			askAggs = append(askAggs, OrderAggregate{Price: price, Quantity: quantity})
		}
	}

	sdCurves, err := NewSDCurves(askAggs, bidAggs)
	if !input.IsOk {
		require.Error(t, err, "SDCurves created, but not expected to")
		return
	}

	require.NoError(t, err, "SDCurves not created, but expected to")
	require.Len(t, sdCurves, len(input.Output), "SDCurves output length mismatch")
	for i := range input.Output {
		price := sdk.NewUint(input.Output[i].P)
		supply := sdk.NewUint(input.Output[i].S)
		demand := sdk.NewUint(input.Output[i].D)

		require.True(t, price.Equal(sdCurves[i].Price), "%d: Price (expected / received): %s / %s", price, sdCurves[i].Price)
		require.True(t, supply.Equal(sdCurves[i].Supply), "%d: Supply (expected / received): %s / %s", supply, sdCurves[i].Supply)
		require.True(t, demand.Equal(sdCurves[i].Demand), "%d: Demand (expected / received): %s / %s", demand, sdCurves[i].Demand)
	}
}

func Test_SDCurves_New(t *testing.T) {
	// zero Bids
	{
		input := SDCurvesInput{
			IsOk: false,
			Aggs: []SDAggInputItem{
				{P: 50, Q: 50, T: SDAggTypeAsk},
			},
		}
		input.Check(t)
	}

	// zero Asks
	{
		input := SDCurvesInput{
			IsOk: false,
			Aggs: []SDAggInputItem{
				{P: 50, Q: 50, T: SDAggTypeBid},
			},
		}
		input.Check(t)
	}

	// check "highest bid price is lower than lowest ask price"
	{
		input := SDCurvesInput{
			IsOk: false,
			Aggs: []SDAggInputItem{
				{P: 100, Q: 50, T: SDAggTypeBid},
				{P: 150, Q: 50, T: SDAggTypeBid},

				{P: 200, Q: 50, T: SDAggTypeAsk},
				{P: 250, Q: 50, T: SDAggTypeAsk},
			},
		}
		input.Check(t)
	}

	// check OK (no zero rows, no fix)
	{
		input := SDCurvesInput{
			IsOk: true,
			Aggs: []SDAggInputItem{
				{P: 100, Q: 100, T: SDAggTypeBid},
				{P: 150, Q: 200, T: SDAggTypeBid},
				{P: 200, Q: 300, T: SDAggTypeBid},

				{P: 100, Q: 300, T: SDAggTypeAsk},
				{P: 150, Q: 200, T: SDAggTypeAsk},
				{P: 200, Q: 100, T: SDAggTypeAsk},
			},
			Output: []SDOutputItem{
				{P: 100, S: 300, D: 100},
				{P: 150, S: 200, D: 200},
				{P: 200, S: 100, D: 300},
			},
		}
		input.Check(t)
	}

	// check OK (with zero rows, with fix)
	{
		input := SDCurvesInput{
			IsOk: true,
			Aggs: []SDAggInputItem{
				{P: 50, Q: 500, T: SDAggTypeBid},
				{P: 100, Q: 400, T: SDAggTypeBid},
				{P: 150, Q: 350, T: SDAggTypeBid},
				{P: 200, Q: 300, T: SDAggTypeBid},
				{P: 250, Q: 150, T: SDAggTypeBid},
				{P: 300, Q: 50, T: SDAggTypeBid},

				{P: 100, Q: 50, T: SDAggTypeAsk},
				{P: 125, Q: 75, T: SDAggTypeAsk},
				{P: 200, Q: 100, T: SDAggTypeAsk},
				{P: 225, Q: 200, T: SDAggTypeAsk},
				{P: 400, Q: 500, T: SDAggTypeAsk},
			},
			Output: []SDOutputItem{
				{P: 50, S: 0, D: 500},
				{P: 100, S: 50, D: 400},
				{P: 125, S: 75, D: 350},
				{P: 150, S: 75, D: 350},
				{P: 200, S: 100, D: 300},
				{P: 225, S: 200, D: 150},
				{P: 250, S: 200, D: 150},
				{P: 300, S: 200, D: 50},
				{P: 400, S: 500, D: 0},
			},
		}
		input.Check(t)
	}
}

type ClearanceStateInput struct {
	Curves         []SDOutputItem
	ClearancePrice uint64
	ClearanceIdx   int
	IsOk           bool
}

func (input ClearanceStateInput) Check(t *testing.T, caseName string) {
	sdCurves := make(SDCurves, 0, len(input.Curves))
	for _, item := range input.Curves {
		sdCurves = append(sdCurves, SDItem{
			Price:  sdk.NewUint(item.P),
			Supply: sdk.NewUint(item.S),
			Demand: sdk.NewUint(item.D),
		})
	}

	fmt.Printf("Case %q:\n%s\n", caseName, sdCurves.Graph())

	state, err := sdCurves.GetClearanceState()
	if !input.IsOk {
		require.Error(t, err, "ClearanceState received, but not expected to")
		return
	}

	require.True(t, input.ClearanceIdx < len(input.Curves))
	require.True(t, input.ClearanceIdx >= 0)

	price := sdk.NewUint(input.ClearancePrice)
	supplyVolumeDec := sdk.NewDecFromBigInt(sdk.NewUint(input.Curves[input.ClearanceIdx].S).BigInt())
	demandVolumeDec := sdk.NewDecFromBigInt(sdk.NewUint(input.Curves[input.ClearanceIdx].D).BigInt())
	proRata := supplyVolumeDec.Quo(demandVolumeDec)
	proRataInvert := sdk.OneDec().Quo(proRata)
	maxBidVolume := sdk.NewDecFromInt(demandVolumeDec.Mul(proRata).RoundInt())
	MaxAskVolume := sdk.NewDecFromInt(supplyVolumeDec.Mul(proRataInvert).RoundInt())

	require.True(t, state.Price.Equal(price), "State: Price (expected / received): %d / %s", input.ClearancePrice, state.Price)
	require.True(t, state.MaxAskVolume.Equal(MaxAskVolume), "State: MaxAskVolume (expected / received): %s / %s", MaxAskVolume, state.MaxAskVolume)
	require.True(t, state.MaxBidVolume.Equal(maxBidVolume), "State: MaxBidVolume (expected / received): %s / %s", maxBidVolume, state.MaxBidVolume)
	require.True(t, state.ProRata.Equal(proRata), "State: ProRata")
	require.True(t, state.ProRataInvert.Equal(proRataInvert), "State: ProRataInvert")
}

func Test_SDCurves_State(t *testing.T) {
	// empty curves
	{
		input := ClearanceStateInput{
			IsOk:   false,
			Curves: nil,
		}
		input.Check(t, "Empty curves")
	}

	// classic case (one cross)
	{
		input := ClearanceStateInput{
			IsOk: true,
			Curves: []SDOutputItem{
				{P: 50, S: 0, D: 100},
				{P: 100, S: 20, D: 90},
				{P: 150, S: 50, D: 80},
				{P: 200, S: 60, D: 70},
				{P: 250, S: 80, D: 10}, // [4] crossing point
				{P: 300, S: 100, D: 0},
			},
			ClearancePrice: 250,
			ClearanceIdx:   4,
		}
		input.Check(t, "One cross")
	}

	// classic case (tunnel with length 2)
	{
		input := ClearanceStateInput{
			IsOk: true,
			Curves: []SDOutputItem{
				{P: 50, S: 0, D: 100},
				{P: 100, S: 20, D: 90},
				{P: 150, S: 50, D: 80},
				{P: 200, S: 60, D: 70},
				{P: 250, S: 80, D: 10}, // [4] tunnel start
				{P: 300, S: 80, D: 10}, // [5] tunnel end
				{P: 350, S: 100, D: 0},
			},
			ClearancePrice: 275,
			ClearanceIdx:   4,
		}
		input.Check(t, "Tunnel 2")
	}

	// classic case (tunnel with length 3)
	{
		input := ClearanceStateInput{
			IsOk: true,
			Curves: []SDOutputItem{
				{P: 50, S: 0, D: 100},
				{P: 100, S: 20, D: 90},
				{P: 150, S: 50, D: 80},
				{P: 200, S: 60, D: 70},
				{P: 250, S: 80, D: 10}, // [4] tunnel start
				{P: 300, S: 80, D: 10},
				{P: 350, S: 80, D: 10}, // [6] tunnel end
				{P: 400, S: 100, D: 0},
			},
			ClearancePrice: 300,
			ClearanceIdx:   4,
		}
		input.Check(t, "Tunnel 3")
	}

	// edge case: no crossing point 1 (min diff)
	{
		input := ClearanceStateInput{
			IsOk: true,
			Curves: []SDOutputItem{
				{P: 50, S: 60, D: 50},
				{P: 100, S: 65, D: 50},
				{P: 150, S: 65, D: 45},
				{P: 200, S: 70, D: 30},
				{P: 250, S: 80, D: 20},
				{P: 300, S: 100, D: 5},
			},
			ClearancePrice: 50,
			ClearanceIdx:   0,
		}
		input.Check(t, "No cross 1: min diff")
	}

	// edge case: no crossing point 2 (zero min diff)
	{
		input := ClearanceStateInput{
			IsOk: true,
			Curves: []SDOutputItem{
				{P: 50, S: 50, D: 50},
				{P: 100, S: 65, D: 50},
				{P: 150, S: 65, D: 45},
				{P: 200, S: 70, D: 30},
				{P: 250, S: 80, D: 20},
				{P: 300, S: 100, D: 5},
			},
			ClearancePrice: 50,
			ClearanceIdx:   0,
		}
		input.Check(t, "No cross 2: zero min diff")
	}

	// edge case: no crossing point 3 (equal min diffs)
	{
		input := ClearanceStateInput{
			IsOk: true,
			Curves: []SDOutputItem{
				{P: 50, S: 100, D: 5},
				{P: 100, S: 80, D: 20},
				{P: 150, S: 60, D: 50},
				{P: 200, S: 60, D: 50},
				{P: 250, S: 80, D: 20},
				{P: 300, S: 100, D: 5},
			},
			ClearancePrice: 150,
			ClearanceIdx:   2,
		}
		input.Check(t, "No cross 3: equal min diffs")
	}

	// edge case: no crossing point 4 (sorting check)
	{
		input := ClearanceStateInput{
			IsOk: true,
			Curves: []SDOutputItem{
				{P: 50, S: 100, D: 5},
				{P: 100, S: 80, D: 20},
				{P: 150, S: 60, D: 50},
				{P: 200, S: 60, D: 55},
				{P: 250, S: 80, D: 20},
				{P: 300, S: 100, D: 5},
			},
			ClearancePrice: 200,
			ClearanceIdx:   3,
		}
		input.Check(t, "No cross 4: sorting check")
	}
}
