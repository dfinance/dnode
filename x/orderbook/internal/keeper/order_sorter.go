package keeper

import orderTypes "github.com/dfinance/dnode/x/order"

type ByPriceAscIDDesc orderTypes.Orders

func (s ByPriceAscIDDesc) Len() int {
	return len(s)
}

func (s ByPriceAscIDDesc) Less(i, j int) bool {
	if s[i].Price.Equal(s[j].Price) {
		return s[i].ID.GT(s[j].ID)
	}

	return s[i].Price.LT(s[j].Price)
}

func (s ByPriceAscIDDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type ByPriceAscIDAsc orderTypes.Orders

func (s ByPriceAscIDAsc) Len() int {
	return len(s)
}

func (s ByPriceAscIDAsc) Less(i, j int) bool {
	if s[i].Price.Equal(s[j].Price) {
		return s[i].ID.LT(s[j].ID)
	}

	return s[i].Price.LT(s[j].Price)
}

func (s ByPriceAscIDAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

