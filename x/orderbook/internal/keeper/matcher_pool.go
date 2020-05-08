package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	orderTypes "github.com/dfinance/dnode/x/order"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
)

type MatcherPool struct {
	logger log.Logger
	pool map[dnTypes.ID]*Matcher
}

func (mp *MatcherPool) AddOrder(order orderTypes.Order) error {
	marketID := order.Market.ID
	matcher, ok := mp.pool[marketID]
	if !ok {
		matcher = NewMatcher(marketID, mp.logger)
		mp.pool[marketID] = matcher
	}

	return matcher.AddOrder(&order)
}

func (mp *MatcherPool) Process() types.MatcherResults {
	results := make(types.MatcherResults, 0, len(mp.pool))

	for marketID, matcher := range mp.pool {
		result, err := matcher.Match()
		if err != nil {
			errMsg := fmt.Sprintf("matcher for marketID %s: %v", marketID.String(), err)
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

func NewMatcherPool(logger log.Logger) MatcherPool {
	return MatcherPool{
		logger: logger,
		pool:   make(map[dnTypes.ID]*Matcher, 0),
	}
}
