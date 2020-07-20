package types

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/olekukonko/tablewriter"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// HistoryItem used to store clearanceState and other meta per block.
// History is preserved per market and if there was a matching.
type HistoryItem struct {
	// MarketID
	MarketID dnTypes.ID `json:"market_id" yaml:"market_id" example:"0" format:"string representation for big.Uint" swaggertype:"string"`
	// Clearance price
	ClearancePrice sdk.Uint `json:"clearance_price" yaml:"clearance_price" swaggertype:"string" example:"100"`
	// Total number of active bid orders
	BidOrdersCount int `json:"bid_orders_count" yaml:"bid_orders_count"`
	// Total number of active ask orders
	AskOrdersCount int `json:"ask_orders_count" yaml:"ask_orders_count"`
	// Clearance bid orders volume
	BidVolume sdk.Uint `json:"bid_volume" yaml:"bid_volume" swaggertype:"string" example:"100"`
	// Clearance ask orders volume
	AskVolume sdk.Uint `json:"ask_volume" yaml:"ask_volume" swaggertype:"string" example:"200"`
	// Matched bid orders volume
	MatchedBidVolume sdk.Uint `json:"matched_bid_volume" yaml:"matched_bid_volume" swaggertype:"string" example:"1000"`
	// Matched ask orders volume
	MatchedAskVolume sdk.Uint `json:"matched_ask_volume" yaml:"matched_ask_volume" swaggertype:"string" example:"2000"`
	// UNIX timestamp [s]
	Timestamp int64 `json:"timestamp" yaml:"timestamp"`
	// Block number
	BlockHeight int64 `json:"block_height" yaml:"block_height"`
}

// Strings returns multi-line text object representation.
func (h HistoryItem) String() string {
	b := strings.Builder{}
	b.WriteString("HistoryItem:\n")
	b.WriteString(fmt.Sprintf("  MarketID:         %s\n", h.MarketID.String()))
	b.WriteString(fmt.Sprintf("  ClearancePrice:   %s\n", h.ClearancePrice.String()))
	b.WriteString(fmt.Sprintf("  BidOrdersCount:   %d\n", h.BidOrdersCount))
	b.WriteString(fmt.Sprintf("  AskOrdersCount:   %d\n", h.AskOrdersCount))
	b.WriteString(fmt.Sprintf("  BidVolume:        %s\n", h.BidVolume.String()))
	b.WriteString(fmt.Sprintf("  AskVolume:        %s\n", h.AskVolume.String()))
	b.WriteString(fmt.Sprintf("  MatchedBidVolume: %s\n", h.MatchedBidVolume.String()))
	b.WriteString(fmt.Sprintf("  MatchedAskVolume: %s\n", h.MatchedAskVolume.String()))
	b.WriteString(fmt.Sprintf("  Timestamp [s]:    %d\n", h.Timestamp))
	b.WriteString(fmt.Sprintf("  BlockHeight:      %d\n", h.BlockHeight))

	return b.String()
}

// TableHeaders returns table headers for multi-line text table object representation.
func (h HistoryItem) TableHeaders() []string {
	headers := []string{
		"H.MarketID",
		"H.ClearancePrice",
		"H.BidOrdersCount",
		"H.AskOrdersCount",
		"H.BidVolume",
		"H.AskVolume",
		"H.MatchedBidVolume",
		"H.MatchedAskVolume",
		"H.Timestamp [s]",
		"H.BlockHeight",
	}

	return headers
}

// TableHeaders returns table rows for multi-line text table object representation.
func (h HistoryItem) TableValues() []string {
	values := []string{
		h.MarketID.String(),
		h.ClearancePrice.String(),
		strconv.FormatInt(int64(h.BidOrdersCount), 10),
		strconv.FormatInt(int64(h.AskOrdersCount), 10),
		h.BidVolume.String(),
		h.AskVolume.String(),
		h.MatchedBidVolume.String(),
		h.MatchedAskVolume.String(),
		time.Unix(h.Timestamp, 0).String(),
		strconv.FormatInt(h.BlockHeight, 10),
	}

	return values
}

func NewHistoryItem(ctx sdk.Context, result MatcherResult) HistoryItem {
	return HistoryItem{
		MarketID:         result.MarketID,
		ClearancePrice:   result.ClearanceState.Price,
		BidOrdersCount:   result.BidOrdersCount,
		AskOrdersCount:   result.AskOrdersCount,
		BidVolume:        sdk.Uint(result.ClearanceState.MaxBidVolume.TruncateInt()),
		AskVolume:        sdk.Uint(result.ClearanceState.MaxAskVolume.TruncateInt()),
		MatchedBidVolume: sdk.Uint(result.MatchedBidVolume.TruncateInt()),
		MatchedAskVolume: sdk.Uint(result.MatchedAskVolume.TruncateInt()),
		Timestamp:        ctx.BlockTime().Unix(),
		BlockHeight:      ctx.BlockHeight(),
	}
}

// HistoryItem slice type.
type HistoryItems []HistoryItem

// Strings returns multi-line text object representation.
func (hi HistoryItems) String() string {
	var buf bytes.Buffer

	t := tablewriter.NewWriter(&buf)
	t.SetHeader(HistoryItem{}.TableHeaders())

	for _, h := range hi {
		t.Append(h.TableValues())
	}
	t.Render()

	return buf.String()
}
