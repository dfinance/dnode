package msgs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func checkExpectedErr(t *testing.T, expectedErr, receivedErr sdk.Error) {
	require.Equal(t, expectedErr.Codespace(), receivedErr.Codespace(), "codeSpace")
	require.Equal(t, expectedErr.Code(), receivedErr.Code(), "code")
}