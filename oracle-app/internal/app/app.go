package app

import (
	"fmt"
	"net"
	"os"
	"os/signal"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"

	"github.com/WingsDao/wings-blockchain/oracle-app/internal/api"
	"github.com/WingsDao/wings-blockchain/oracle-app/internal/exchange"
	_ "github.com/WingsDao/wings-blockchain/oracle-app/internal/exchange/binance"
)

type Config struct {
	ChainID    string
	Mnemonic   string
	APIAddress string
	Gas        uint64
	Fees       string
	Logger     *logrus.Logger
	Assets     map[string][]exchange.Asset
}

type OracleApp struct {
	config    *Config
	stopCh    chan struct{}
	tickersCh chan exchange.Ticker
	cl        *api.Client
}

func NewOracleApp(c *Config) (*OracleApp, error) {
	fees, err := sdk.ParseCoins(c.Fees)
	if err != nil {
		return nil, err
	}

	_, _, err = net.SplitHostPort(c.APIAddress)
	if err != nil {
		return nil, err
	}

	apiCl, err := api.NewClient(c.Mnemonic, c.ChainID, c.APIAddress, fees)
	if err != nil {
		return nil, err
	}
	return &OracleApp{
		config:    c,
		stopCh:    make(chan struct{}),
		tickersCh: make(chan exchange.Ticker, 100),
		cl:        apiCl,
	}, nil
}

func (a *OracleApp) Start() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	if err := a.lisrenUpdates(); err != nil {
		return err
	}
	<-c
	close(c)
	close(a.stopCh)
	return nil
}

func (a *OracleApp) lisrenUpdates() error {
	for name, subscriber := range exchange.Exchanges() {
		assets, ok := a.config.Assets[name]
		if !ok {
			return fmt.Errorf("%s: assets config not found", name)
		}
		for _, asset := range assets {
			err := subscriber.Subscribe(exchange.NewAsset(asset.Code, asset.Pair), a.tickersCh)
			if err != nil {
				return fmt.Errorf("%s: subscribe error: %s", name, err)
			}
		}
	}
	go func() {
		for {
			ticker, ok := <-a.tickersCh
			if !ok {
				return
			}
			err := a.cl.PostPrice(ticker)
			if err != nil {
				logrus.Errorf("error while post ticker [%s]: %s", ticker, err)
			} else {
				logrus.Infof("posted price from [%s] for [%s]", ticker.Exchange, ticker)
			}
		}
	}()

	return nil
}
