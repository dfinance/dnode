package msgs

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	ethAddress = "0x82A978B3f5962A5b0957d9ee9eEf472EE55B42F1"
)

func checkExpectedErr(t *testing.T, expectedErr, receivedErr sdk.Error) {
	require.Equal(t, expectedErr.Codespace(), receivedErr.Codespace(), "codeSpace")
	require.Equal(t, expectedErr.Code(), receivedErr.Code(), "code")
}
