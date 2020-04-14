package tests

import (
	"net"
	"testing"
	"time"

	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func PingTcpAddress(address string) error {
	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}

func CheckExpectedErr(t *testing.T, expectedErr, receivedErr error) {
	require.NotNil(t, receivedErr, "receivedErr")

	expectedSdkErr, ok := expectedErr.(*sdkErrors.Error)
	require.True(t, ok, "not a SDK error: %T", expectedErr)

	require.True(t, expectedSdkErr.Is(receivedErr), "receivedErr / expectedErr: %v / %v", receivedErr, expectedErr)
}
