package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	orderTypes "github.com/dfinance/dnode/x/order"
)

// MatcherResult stores matcher results.
type MatcherResult struct {
	// PQCurve crossing point data
	ClearanceState ClearanceState
	// Sum of matched bid orders volume
	MatchedBidVolume sdk.Dec
	// Sum of matched ask orders volume
	MatchedAskVolume sdk.Dec
	// Fully / partially filled orders with some meta
	OrderFills orderTypes.OrderFills
}

// Strings returns multi-line text object representation.
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

// MatcherResult slice type.
type MatcherResults []MatcherResult
