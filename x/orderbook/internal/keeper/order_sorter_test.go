// +build unit

package keeper

import (
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orders"
)

func TestOBKeeper_Sorter(t *testing.T) {
	t.Parallel()

	orders := orders.Orders{
		orders.Order{ID: dnTypes.NewIDFromUint64(1), Price: sdk.NewUint(100)},
		orders.Order{ID: dnTypes.NewIDFromUint64(0), Price: sdk.NewUint(50)},
		orders.Order{ID: dnTypes.NewIDFromUint64(2), Price: sdk.NewUint(50)},
		orders.Order{ID: dnTypes.NewIDFromUint64(6), Price: sdk.NewUint(200)},
		orders.Order{ID: dnTypes.NewIDFromUint64(5), Price: sdk.NewUint(100)},
		orders.Order{ID: dnTypes.NewIDFromUint64(7), Price: sdk.NewUint(25)},
	}

	sort.Sort(ByPriceAscIDDesc(orders))
	isSortedByPriceAscIDDesc := sort.SliceIsSorted(orders, func(i, j int) bool {
		if orders[i].Price.Equal(orders[j].Price) {
			return orders[i].ID.GTE(orders[j].ID)
		}

		return orders[i].Price.LTE(orders[j].Price)
	})
	require.True(t, isSortedByPriceAscIDDesc)

	sort.Sort(ByPriceAscIDAsc(orders))
	isSortedByPriceAscIDAsc := sort.SliceIsSorted(orders, func(i, j int) bool {
		if orders[i].Price.Equal(orders[j].Price) {
			return orders[i].ID.LTE(orders[j].ID)
		}

		return orders[i].Price.LTE(orders[j].Price)
	})
	require.True(t, isSortedByPriceAscIDAsc)
}
