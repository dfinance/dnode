package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/x/orderbook/internal/types"
	orderTypes "github.com/dfinance/dnode/x/orders"
)

// MatcherPool objects stores matchers for market IDs.
type MatcherPool struct {
	logger log.Logger
	pool   map[string]*Matcher
}

// AddOrder adds order to the corresponding matcher (by marketID).
func (mp *MatcherPool) AddOrder(order orderTypes.Order) error {
	marketID := order.Market.ID
	matcher, ok := mp.pool[marketID.String()]
	if !ok {
		matcher = NewMatcher(marketID, mp.logger)
		mp.pool[marketID.String()] = matcher
	}

	return matcher.AddOrder(&order)
}

// Process executes every pool matcher and combines the results.
// Panics on internal errors, otherwise just logs.
func (mp *MatcherPool) Process() types.MatcherResults {
	results := make(types.MatcherResults, 0, len(mp.pool))

	for marketID, matcher := range mp.pool {
		result, err := matcher.Match()
		if err != nil {
			errMsg := fmt.Sprintf("matcher for marketID %s: %v", marketID, err)
			if types.ErrInternal.Is(err) {
				panic(errMsg)
			} else {
				mp.logger.Info(errMsg)
			}

			continue
		}

		results = append(results, result)
	}

	return results
}

// NewMatcherPool creates a new MatcherPool object.
func NewMatcherPool(logger log.Logger) MatcherPool {
	return MatcherPool{
		logger: logger,
		pool:   make(map[string]*Matcher),
	}
}
