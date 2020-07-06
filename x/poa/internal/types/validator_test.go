// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestPoa_Validator_Validate(t *testing.T) {
	t.Parallel()

	// ok
	{
		sdkAddr := sdk.AccAddress("addr1")
		ethAddr := "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7d"
		v := NewValidator(sdkAddr, ethAddr)
		require.NoError(t, v.Validate())
	}

	// fail: empty address
	{
		sdkAddr := sdk.AccAddress("")
		ethAddr := "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7d"
		v := NewValidator(sdkAddr, ethAddr)
		require.Error(t, v.Validate())
	}

	// fail: empty ethAddress
	{
		sdkAddr := sdk.AccAddress("addr1")
		ethAddr := ""
		v := NewValidator(sdkAddr, ethAddr)
		require.Error(t, v.Validate())
	}

	// fail: invalid ethAddress
	{
		sdkAddr := sdk.AccAddress("addr1")
		ethAddr := "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7"
		v := NewValidator(sdkAddr, ethAddr)
		require.Error(t, v.Validate())
	}
}
