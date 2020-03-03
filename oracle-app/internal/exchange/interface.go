package exchange

type Subscriber interface {
	Subscribe(Asset, chan Ticker) error
	//Name() string
}
