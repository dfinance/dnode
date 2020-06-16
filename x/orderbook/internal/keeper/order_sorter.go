package keeper

import (
	orderTypes "github.com/dfinance/dnode/x/orders"
)

// ByPriceAscIDDesc is a type wrapper used to sort orders slice by Price ASC (1st priority) and ID DESC (2nd priority).
type ByPriceAscIDDesc orderTypes.Orders

// Implements sort.Interface.
func (s ByPriceAscIDDesc) Len() int {
	return len(s)
}

// Implements sort.Interface.
func (s ByPriceAscIDDesc) Less(i, j int) bool {
	if s[i].Price.Equal(s[j].Price) {
		return s[i].ID.GT(s[j].ID)
	}

	return s[i].Price.LT(s[j].Price)
}

// Implements sort.Interface.
func (s ByPriceAscIDDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// ByPriceAscIDAsc is a type wrapper used to sort orders slice by Price ASC (1st priority) and ID ASC (2nd priority).
type ByPriceAscIDAsc orderTypes.Orders

// Implements sort.Interface.
func (s ByPriceAscIDAsc) Len() int {
	return len(s)
}

// Implements sort.Interface.
func (s ByPriceAscIDAsc) Less(i, j int) bool {
	if s[i].Price.Equal(s[j].Price) {
		return s[i].ID.LT(s[j].ID)
	}

	return s[i].Price.LT(s[j].Price)
}

// Implements sort.Interface.
func (s ByPriceAscIDAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

