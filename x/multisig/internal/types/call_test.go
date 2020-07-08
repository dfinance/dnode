// +build unit

package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

type MockMsMsg struct {
	msRoute string
	msType  string
	isValid bool
}

func (m MockMsMsg) Route() string { return m.msRoute }
func (m MockMsMsg) Type() string  { return m.msType }
func (m MockMsMsg) ValidateBasic() error {
	if !m.isValid {
		return fmt.Errorf("some error")
	}
	return nil
}
func NewMockMsMsg(msRoute, msType string, isValid bool) MockMsMsg {
	return MockMsMsg{msRoute: msRoute, msType: msType, isValid: isValid}
}

func NewOkMockMsMsg() MockMsMsg           { return NewMockMsMsg("route", "type", true) }
func NewInvalidRouteMockMsMsg() MockMsMsg { return NewMockMsMsg("", "type", true) }
func NewInvalidTypeMockMsMsg() MockMsMsg  { return NewMockMsMsg("route", "", true) }
func NewInvalidMockMsMsg() MockMsMsg      { return NewMockMsMsg("route", "type", false) }

// Check new call validation.
func TestMS_NewCall(t *testing.T) {
	t.Parallel()

	// ok
	{
		addr := sdk.AccAddress("addr1")
		_, err := NewCall(dnTypes.NewIDFromUint64(0), "unique", NewOkMockMsMsg(), 0, addr)
		require.NoError(t, err)
	}

	// fail: nil msg
	{
		addr := sdk.AccAddress("addr1")
		_, err := NewCall(dnTypes.NewIDFromUint64(0), "unique", nil, 0, addr)
		require.Error(t, err)
	}

	// fail: invalid msg: route
	{
		addr := sdk.AccAddress("addr1")
		_, err := NewCall(dnTypes.NewIDFromUint64(0), "unique", NewInvalidRouteMockMsMsg(), 0, addr)
		require.Error(t, err)
	}

	// fail: invalid msg: type
	{
		addr := sdk.AccAddress("addr1")
		_, err := NewCall(dnTypes.NewIDFromUint64(0), "unique", NewInvalidTypeMockMsMsg(), 0, addr)
		require.Error(t, err)
	}

	// fail: invalid ID
	{
		addr := sdk.AccAddress("addr1")
		_, err := NewCall(dnTypes.ID{}, "unique", NewOkMockMsMsg(), 0, addr)
		require.Error(t, err)
	}

	// fail: invalid uniqueID
	{
		addr := sdk.AccAddress("addr1")
		_, err := NewCall(dnTypes.NewIDFromUint64(0), "", NewOkMockMsMsg(), 0, addr)
		require.Error(t, err)
	}

	// fail: invalid creator
	{
		addr := sdk.AccAddress{}
		_, err := NewCall(dnTypes.NewIDFromUint64(0), "", NewOkMockMsMsg(), 0, addr)
		require.Error(t, err)
	}
}
