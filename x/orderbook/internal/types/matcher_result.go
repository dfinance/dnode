package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	orderTypes "github.com/dfinance/dnode/x/order"
)

type MatcherResult struct {
	ClearanceState   ClearanceState
	MatchedBidVolume sdk.Dec
	MatchedAskVolume sdk.Dec
	OrderFills       orderTypes.OrderFills
}

func (r MatcherResult) String() string {
	b := strings.Builder{}
	b.WriteString("MatcherResult:\n")
	b.WriteString(fmt.Sprintf("  MatchedBidVolume: %s\n", r.MatchedBidVolume.String()))
	b.WriteString(fmt.Sprintf("  MatchedAskVolume: %s\n", r.MatchedAskVolume.String()))
	b.WriteString(r.ClearanceState.String())
	b.WriteString("OrderFills:\n")
	b.WriteString(r.OrderFills.String())

	return b.String()
}

type MatcherResults []MatcherResult
