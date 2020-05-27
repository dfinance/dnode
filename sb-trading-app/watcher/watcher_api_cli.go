package watcher

import (
	"fmt"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	marketTypes "github.com/dfinance/dnode/x/markets"
)

type Api interface {
	AddMarket(baseDenom, quoteDenom string) error
}

type ApiCli struct {
	tester               *cliTester.CLITester
	marketCreatorAccName string
}

func (a *ApiCli) AddMarket(baseDenom, quoteDenom string) (*marketTypes.Market, error) {
	r := a.tester.TxMarketsAdd(a.marketCreatorAccName, baseDenom, quoteDenom)
	if _, err := r.Execute(); err != nil {
		return nil, fmt.Errorf("AddMarket: creating: %w", err)
	}

	q, markets := a.tester.QueryMarketsList(-1, -1, &baseDenom, &quoteDenom)
	if _, err := q.Execute(); err != nil {
		return nil, fmt.Errorf("AddMarket: markets query: %w", err)
	}
	if markets == nil || len(*markets) != 1 {
		return nil, fmt.Errorf("AddMarket: markets query: invalid merkets len")
	}
	market := (*markets)[0]

	return &market, nil
}

func NewApiCli(tester *cliTester.CLITester, marketCreatorAccName string) *ApiCli {
	return &ApiCli{
		tester:               tester,
		marketCreatorAccName: marketCreatorAccName,
	}
}
