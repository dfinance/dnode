package keeper

import (
	"bytes"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/olekukonko/tablewriter"

	"github.com/dfinance/dnode/x/orderbook/internal/types"
)

// Supply-demand curves point holding aggregated supply and demand quantities for price.
type SDItem struct {
	Price  sdk.Uint
	Supply sdk.Uint
	Demand sdk.Uint
}

// SupplyDemandBalance compares point supply and demand.
func (i SDItem) SupplyDemandBalance() int {
	if i.Supply.GT(i.Demand) {
		return 1
	}
	if i.Supply.Equal(i.Demand) {
		return 0
	}

	return -1
}

// SDCurves object stores all Supply-demand curves points.
type SDCurves []SDItem

// Strings returns multi-line text object representation.
func (c *SDCurves) String() string {
	var buf bytes.Buffer

	t := tablewriter.NewWriter(&buf)
	t.SetHeader([]string{
		"PQ.Price",
		"PQ.Supply",
		"PQ.Demand",
	})

	for _, i := range *c {
		t.Append([]string{
			i.Price.String(),
			i.Supply.String(),
			i.Demand.String(),
		})
	}
	t.Render()

	return buf.String()
}

// NewSDCurves creates a new SDCurves object mergind ask/bid aggregates.
func NewSDCurves(askAggs, bidAggs OrderAggregates) (SDCurves, error) {
	// check if curves can be obtained
	if len(askAggs) == 0 || len(bidAggs) == 0 {
		return SDCurves{}, fmt.Errorf("ask / bid orders are empty")
	}

	// check inputs (aggregates must be properly sorted)
	// TODO: remove that time wasting checks once we are sure aggregates build is correct
	askAggsSorted := sort.SliceIsSorted(askAggs, func(i, j int) bool {
		return askAggs[i].Price.LT(askAggs[j].Price) && askAggs[i].Quantity.LTE(askAggs[j].Quantity)
	})
	if !askAggsSorted {
		return SDCurves{}, sdkErrors.Wrap(types.ErrInternal, "askAggs not sorted")
	}

	bidAggsSorted := sort.SliceIsSorted(bidAggs, func(i, j int) bool {
		return bidAggs[i].Price.LT(bidAggs[j].Price) && bidAggs[i].Quantity.GTE(bidAggs[j].Quantity)
	})
	if !bidAggsSorted {
		return SDCurves{}, sdkErrors.Wrap(types.ErrInternal, "bidAggs not sorted")
	}

	// check if clearance price can be found
	if bidAggs[len(bidAggs)-1].Price.LT(askAggs[0].Price) {
		return SDCurves{}, fmt.Errorf("highest bid price is lower than lowest ask price")
	}

	// merge bid/ask aggregates inputs
	c := SDCurves{}
	c.addAskOrders(askAggs)
	c.addBidOrders(bidAggs)
	c.fixZeroSupplyDemand()

	return c, nil
}

// GetClearanceState processes the SDCurves searching for the best crossing point.
// Result is the clearance price and adjusted (by ProRate coefficient) max supply/demand amounts that are
// used to fill orders.
func (c *SDCurves) GetClearanceState() (retState types.ClearanceState, retErr error) {
	// check input
	if len(*c) == 0 {
		retErr = fmt.Errorf("SDCurves is empty")
		return
	}

	crossPoint := c.getCrossPoint()
	// check is the crossing point is valid
	if crossPoint.Price.IsZero() {
		retErr = sdkErrors.Wrap(types.ErrInternal, "crossPoint.Price: empty")
		return
	}
	if crossPoint.Supply.IsZero() {
		retErr = sdkErrors.Wrap(types.ErrInternal, "crossPoint.Supply: empty")
		return
	}
	if crossPoint.Demand.IsZero() {
		retErr = sdkErrors.Wrap(types.ErrInternal, "crossPoint.Demand: empty")
		return
	}

	// convert demand/supply to sdk.Dec for better accuracy
	demandDec, supplyDec := sdk.NewDecFromBigInt(crossPoint.Demand.BigInt()), sdk.NewDecFromBigInt(crossPoint.Supply.BigInt())
	// build the result
	retState.Price = crossPoint.Price
	retState.ProRata = supplyDec.Quo(demandDec)
	retState.ProRataInvert = sdk.OneDec().Quo(retState.ProRata)
	retState.MaxBidVolume = sdk.NewDecFromInt(demandDec.Mul(retState.ProRata).RoundInt())
	retState.MaxAskVolume = sdk.NewDecFromInt(supplyDec.Mul(retState.ProRataInvert).RoundInt())

	return
}

// getCrossPoint searches for the crossing point.
// Crossing point might not be found: other point is picked in that case (edge cases).
func (c *SDCurves) getCrossPoint() SDItem {
	// crossPointIdx is an index of the last found crossing point
	// clearancePrice is the last calculated clearance price
	crossPointIdx, clearancePrice := -1, sdk.ZeroUint()
	cLen := len(*c)

	// edge-case 0: first point has non-zero equal Supply and Demand
	// it's a crossing point, but left SupplyDemandBalance can't be calculated (idx == -1)
	if firstItem := &(*c)[0]; firstItem.Supply.Equal(firstItem.Demand) {
		if !firstItem.Supply.IsZero() && !firstItem.Demand.IsZero() {
			return SDItem{
				Price:  firstItem.Price,
				Supply: firstItem.Supply,
				Demand: firstItem.Demand,
			}
		}
	}

	for i := 1; i < cLen; i++ {
		curItem, prevItem := &(*c)[i], &(*c)[i-1]

		// cross point is defined by previous and current point having different supply/demand relations
		if prevItem.SupplyDemandBalance() != curItem.SupplyDemandBalance() {
			crossPointIdx, clearancePrice = i, curItem.Price

			// check if next points are equal to the found one ("corridor")
			leftCrossPoint, rightCrossPoint := curItem, (*SDItem)(nil)
			for j := i + 1; j < cLen; j++ {
				rightItem := &(*c)[j]
				if !leftCrossPoint.Supply.Equal(rightItem.Supply) || !leftCrossPoint.Demand.Equal(rightItem.Demand) {
					break
				}
				rightCrossPoint = rightItem
			}
			// if the "corridor" was found average the clearance price
			if rightCrossPoint != nil {
				clearancePrice = leftCrossPoint.Price.Add(rightCrossPoint.Price).QuoUint64(2)
			}

			break
		}
	}

	if crossPointIdx != -1 {
		// the crossing point was found
		curItem, prevItem := &(*c)[crossPointIdx], &(*c)[crossPointIdx-1]

		if !curItem.Supply.IsZero() && !curItem.Demand.IsZero() {
			// edge-case 1a: crossing point has volumes
			return SDItem{
				Price:  clearancePrice,
				Supply: curItem.Supply,
				Demand: curItem.Demand,
			}
		}

		if !prevItem.Supply.IsZero() && !prevItem.Demand.IsZero() {
			// edge-case 1b: prev to the crossing point has volumes
			return SDItem{
				Price:  prevItem.Price,
				Supply: prevItem.Supply,
				Demand: prevItem.Demand,
			}
		}
	}

	// the crossing point wasn't found

	// edge-case 2 (option a): find point with the minimal |Supply - Demand| diff and non-zero Supply and Demand
	// Supply / Demand can't be zero as it would cause "div 0" error
	cSorted := make(SDCurves, cLen)
	copy(cSorted, *c)
	sort.Slice(cSorted, func(i, j int) bool {
		leftItem, rightItem := &cSorted[i], &cSorted[j]

		var leftDiff, rightDiff sdk.Uint
		if leftItem.Supply.GT(leftItem.Demand) {
			leftDiff = leftItem.Supply.Sub(leftItem.Demand)
		} else {
			leftDiff = leftItem.Demand.Sub(leftItem.Supply)
		}
		if rightItem.Supply.GT(rightItem.Demand) {
			rightDiff = rightItem.Supply.Sub(rightItem.Demand)
		} else {
			rightDiff = rightItem.Demand.Sub(rightItem.Supply)
		}

		return leftDiff.LT(rightDiff)
	})
	for i := 0; i < len(cSorted); i++ {
		item := &cSorted[i]
		if !item.Supply.IsZero() && !item.Demand.IsZero() {
			return SDItem{
				Price:  item.Price,
				Supply: item.Supply,
				Demand: item.Demand,
			}
		}
	}

	// edge-case 2 (option b): find first point with non-zero Supply and Demand (search from the middle)
	// Supply / Demand can't be zero as it would cause "div 0" error
	//middleIdx := cLen / 2
	//for rightIdx := middleIdx; rightIdx < cLen; rightIdx++ {
	//	rightItem := &(*c)[rightIdx]
	//
	//	var leftItem *SDItem
	//	leftIdx := middleIdx - (rightIdx - middleIdx)
	//	if leftIdx >= 0 {
	//		leftItem = &(*c)[leftIdx]
	//	}
	//
	//	if !rightItem.Supply.IsZero() && !rightItem.Demand.IsZero() {
	//		return SDItem{
	//			Price:  rightItem.Price,
	//			Supply: rightItem.Supply,
	//			Demand: rightItem.Demand,
	//		}
	//	}
	//
	//	if leftItem != nil && !leftItem.Supply.IsZero() && !leftItem.Demand.IsZero() {
	//		return SDItem{
	//			Price:  leftItem.Price,
	//			Supply: leftItem.Supply,
	//			Demand: leftItem.Demand,
	//		}
	//	}
	//}

	// edge-case 3: can't happen if the lowest ask price is higher then the highest bid price
	return SDItem{
		Price:  sdk.ZeroUint(),
		Supply: sdk.ZeroUint(),
		Demand: sdk.ZeroUint(),
	}
}

// addAskOrders merges ask aggregates into SDCurve.
func (c *SDCurves) addAskOrders(aggs OrderAggregates) {
	for i := 0; i < len(aggs); i++ {
		agg := &aggs[i]

		*c = append(*c, SDItem{
			Price:  agg.Price,
			Supply: agg.Quantity,
			Demand: sdk.ZeroUint(),
		})
	}
}

// addAskOrders merges bid aggregates into SDCurve.
// Contract: ask aggregates must be merged first.
func (c *SDCurves) addBidOrders(aggs OrderAggregates) {
	for i := 0; i < len(aggs); i++ {
		agg := &aggs[i]

		gtePriceIdx := sort.Search(len(*c), func(i int) bool {
			return (*c)[i].Price.GTE(agg.Price)
		})

		if gtePriceIdx != len(*c) && (*c)[gtePriceIdx].Price.Equal(agg.Price) {
			(*c)[gtePriceIdx].Demand = agg.Quantity
			continue
		}

		*c = append(*c, SDItem{})
		copy((*c)[gtePriceIdx+1:], (*c)[gtePriceIdx:])
		(*c)[gtePriceIdx] = SDItem{
			Price:  agg.Price,
			Supply: sdk.ZeroUint(),
			Demand: agg.Quantity,
		}
	}
}

// fixZeroSupplyDemand sets empty supply/demand for inserted rows.
// Iterates over SDCurves in both directions copying quantity from the previous point.
func (c *SDCurves) fixZeroSupplyDemand() {
	askIdx, askFillItem := 1, &(*c)[0]
	bidIdx, bidFillItem := len(*c)-2, &(*c)[len(*c)-1]

	for i := 1; i < len(*c); i++ {
		if (*c)[askIdx].Supply.IsZero() {
			(*c)[askIdx].Supply = (*askFillItem).Supply
		} else {
			askFillItem = &(*c)[askIdx]
		}
		askIdx++

		if (*c)[bidIdx].Demand.IsZero() {
			(*c)[bidIdx].Demand = (*bidFillItem).Demand
		} else {
			bidFillItem = &(*c)[bidIdx]
		}
		bidIdx--
	}
}
