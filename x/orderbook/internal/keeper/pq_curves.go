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

type PQItem struct {
	Price  sdk.Uint
	Supply sdk.Uint
	Demand sdk.Uint
}

func (i PQItem) SupplyDemandBalance() int {
	if i.Supply.GT(i.Demand) {
		return 1
	}
	if i.Supply.Equal(i.Demand) {
		return 0
	}

	return -1
}

type PQCurves []PQItem

func (c *PQCurves) String() string {
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

	return string(buf.Bytes())
}

func NewPQCurves(askAggs, bidAggs OrderAggregates) (PQCurves, error) {
	if len(askAggs) == 0 || len(bidAggs) == 0 {
		return PQCurves{}, fmt.Errorf("ask / bid orders are empty")
	}

	askAggsSorted := sort.SliceIsSorted(askAggs, func(i, j int) bool {
		return askAggs[i].Price.LT(askAggs[j].Price) && askAggs[i].Quantity.LTE(askAggs[j].Quantity)
	})
	if !askAggsSorted {
		return PQCurves{}, sdkErrors.Wrap(types.ErrInternal, "askAggs not sorted")
	}

	bidAggsSorted := sort.SliceIsSorted(bidAggs, func(i, j int) bool {
		return bidAggs[i].Price.LT(bidAggs[j].Price) && bidAggs[i].Quantity.GTE(bidAggs[j].Quantity)
	})
	if !bidAggsSorted {
		return PQCurves{}, sdkErrors.Wrap(types.ErrInternal, "bidAggs not sorted")
	}

	if bidAggs[len(bidAggs)-1].Price.LT(askAggs[0].Price) {
		return PQCurves{}, sdkErrors.Wrap(types.ErrInternal, "highest bid price is lower than lowest ask price")
	}

	c := PQCurves{}
	c.addAskOrders(askAggs)
	c.addBidOrders(bidAggs)
	c.fixZeroSupplyDemand()

	return c, nil
}

func (c *PQCurves) GetClearanceState() (retState types.ClearanceState, retErr error) {
	if len(*c) == 0 {
		retErr = fmt.Errorf("PQCurves is empty")
		return
	}

	crossPoint := c.getCrossPoint()
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

	demandDec, supplyDec := sdk.NewDecFromBigInt(crossPoint.Demand.BigInt()), sdk.NewDecFromBigInt(crossPoint.Supply.BigInt())

	retState.Price = crossPoint.Price
	retState.ProRata = supplyDec.Quo(demandDec)
	retState.ProRataInvert = sdk.OneDec().Quo(retState.ProRata)
	retState.MaxBidVolume = sdk.NewDecFromInt(demandDec.Mul(retState.ProRata).RoundInt())
	retState.MaxAskVolume = sdk.NewDecFromInt(supplyDec.Mul(retState.ProRataInvert).RoundInt())

	return
}

func (c *PQCurves) getCrossPoint() PQItem {
	crossPointIdx, clearancePrice := -1, sdk.ZeroUint()
	cLen := len(*c)

	for i := 1; i < cLen; i++ {
		curItem, prevItem := &(*c)[i], &(*c)[i-1]

		if prevItem.SupplyDemandBalance() != curItem.SupplyDemandBalance() {
			crossPointIdx, clearancePrice = i, curItem.Price

			leftCrossPoint, rightCrossPoint := curItem, (*PQItem)(nil)
			for j := i + 1; j < cLen; j++ {
				rightItem := &(*c)[j]
				if !leftCrossPoint.Supply.Equal(rightItem.Supply) || !leftCrossPoint.Demand.Equal(rightItem.Demand) {
					break
				}
				rightCrossPoint = rightItem
			}
			if rightCrossPoint != nil {
				clearancePrice = leftCrossPoint.Price.Add(rightCrossPoint.Price).QuoUint64(2)
			}

			break
		}
	}

	if crossPointIdx != -1 {
		curItem, prevItem := &(*c)[crossPointIdx], &(*c)[crossPointIdx-1]

		if !prevItem.Supply.IsZero() {
			return PQItem{
				Price:  clearancePrice,
				Supply: curItem.Supply,
				Demand: curItem.Demand,
			}
		}

		if !curItem.Supply.IsZero() && !curItem.Demand.IsZero() {
			return PQItem{
				Price:  curItem.Price,
				Supply: curItem.Supply,
				Demand: curItem.Demand,
			}
		}
	}

	if cLen > 1 {
		lastItem := &(*c)[cLen-1]
		return PQItem{
			Price:  lastItem.Price,
			Supply: lastItem.Supply,
			Demand: lastItem.Demand,
		}
	}

	firstItem := &(*c)[0]
	return PQItem{
		Price:  firstItem.Price,
		Supply: firstItem.Supply,
		Demand: firstItem.Demand,
	}
}

func (c *PQCurves) addAskOrders(aggs OrderAggregates) {
	for i := 0; i < len(aggs); i++ {
		agg := &aggs[i]

		*c = append(*c, PQItem{
			Price:  agg.Price,
			Supply: agg.Quantity,
			Demand: sdk.ZeroUint(),
		})
	}
}

func (c *PQCurves) addBidOrders(aggs OrderAggregates) {
	for i := 0; i < len(aggs); i++ {
		agg := &aggs[i]

		gtePriceIdx := sort.Search(len(*c), func(i int) bool {
			return (*c)[i].Price.GTE(agg.Price)
		})

		if (*c)[gtePriceIdx].Price.Equal(agg.Price) {
			(*c)[gtePriceIdx].Demand = agg.Quantity
			continue
		}

		*c = append(*c, PQItem{})
		copy((*c)[gtePriceIdx+1:], (*c)[gtePriceIdx:])
		(*c)[gtePriceIdx] = PQItem{
			Price:  agg.Price,
			Supply: sdk.ZeroUint(),
			Demand: agg.Quantity,
		}
	}
}

func (c *PQCurves) fixZeroSupplyDemand() {
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
