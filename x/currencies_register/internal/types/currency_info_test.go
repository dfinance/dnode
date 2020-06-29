// +build unit

package types

import (
	"encoding/hex"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/common_vm"
)

func TestNewCurrencyInfo(t *testing.T) {
	// Parameters.
	denom := "dfi"
	decimals := 18
	isToken := true
	owner := common_vm.StdLibAddress
	totalSupply := sdk.NewInt(1000000)

	// New currency info.
	currInfo, err := NewCurrencyInfo([]byte(denom), uint8(decimals), isToken, owner, totalSupply.BigInt())
	require.NoError(t, err)

	require.EqualValues(t, denom, currInfo.Denom)
	require.EqualValues(t, decimals, currInfo.Decimals)
	require.EqualValues(t, owner, currInfo.Owner)
	require.EqualValues(t, totalSupply.BigInt().String(), currInfo.TotalSupply.String())

	// Expect error with wrong address.
	_, err = NewCurrencyInfo([]byte(denom), uint8(decimals), isToken, make([]byte, 32), totalSupply.BigInt())
	require.Errorf(t, err, "length of owner address is not equal to address length: %d / %d", len(owner), common_vm.VMAddressLength)
}

// Test String().
func TestCurrencyInfo_String(t *testing.T) {
	denom := "dfi"
	decimals := 18
	isToken := true
	owner := common_vm.StdLibAddress
	totalSupply := sdk.NewInt(1000000)

	currInfo, err := NewCurrencyInfo([]byte(denom), uint8(decimals), isToken, owner, totalSupply.BigInt())
	require.NoError(t, err)

	currStr := fmt.Sprintf("Currency: %s\n"+
		"\tDecimals: %d\n"+
		"\tIs Token: %t\n"+
		"\tOwner:    %s\n"+
		"\tTotal supply: %s\n",
		denom, decimals, isToken,
		hex.EncodeToString(owner), totalSupply.String())

	require.Equal(t, currStr, currInfo.String())
}
