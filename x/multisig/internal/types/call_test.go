// +build unit

package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

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

// Call genesis validation.
func TestMS_Call_Valid(t *testing.T) {
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// fail: ID
	{
		call := Call{
			ID: dnTypes.ID{},
		}
		require.Error(t, call.Valid(-1))
	}
	// fail: UniqueID
	{
		call := Call{
			ID: dnTypes.NewZeroID(),
		}
		require.Error(t, call.Valid(-1))
	}
	// fail: Creator
	{
		call := Call{
			ID:       dnTypes.NewZeroID(),
			UniqueID: "unique",
		}
		require.Error(t, call.Valid(-1))
	}
	// fail: c.Approved && !(c.Executed || c.Failed)
	{
		call := Call{
			ID:       dnTypes.NewZeroID(),
			UniqueID: "unique",
			Creator:  addr,
			Approved: true,
		}
		require.Error(t, call.Valid(-1))
	}
	// fail: c.Rejected && (c.Approved || c.Executed || c.Failed)
	{
		call := Call{
			ID:       dnTypes.NewZeroID(),
			UniqueID: "unique",
			Creator:  addr,
			Approved: true,
			Rejected: true,
		}
		require.Error(t, call.Valid(-1))
	}
	// fail: Msg nil
	{
		call := Call{
			ID:       dnTypes.NewZeroID(),
			UniqueID: "unique",
			Creator:  addr,
			Approved: true,
			Executed: true,
		}
		require.Error(t, call.Valid(-1))
	}
	// fail: MsgRoute empty
	{
		call := Call{
			ID:       dnTypes.NewZeroID(),
			UniqueID: "unique",
			Creator:  addr,
			Approved: true,
			Executed: true,
			Msg:      NewMockMsMsg("route", "type", true),
		}
		require.Error(t, call.Valid(-1))
	}
	// fail: MsgType empty
	{
		call := Call{
			ID:       dnTypes.NewZeroID(),
			UniqueID: "unique",
			Creator:  addr,
			Approved: true,
			Executed: true,
			Msg:      NewMockMsMsg("route", "type", true),
			MsgRoute: "route",
		}
		require.Error(t, call.Valid(-1))
	}
	// fail: Height
	{
		call := Call{
			ID:       dnTypes.NewZeroID(),
			UniqueID: "unique",
			Creator:  addr,
			Approved: true,
			Executed: true,
			Msg:      NewMockMsMsg("route", "type", true),
			MsgRoute: "route",
			MsgType:  "type",
			Height:   -1,
		}
		require.Error(t, call.Valid(-1))
	}
	// fail: Height > CurBlockHeight
	{
		call := Call{
			ID:       dnTypes.NewZeroID(),
			UniqueID: "unique",
			Creator:  addr,
			Approved: true,
			Executed: true,
			Msg:      NewMockMsMsg("route", "type", true),
			MsgRoute: "route",
			MsgType:  "type",
			Height:   100,
		}
		require.Error(t, call.Valid(1))
	}
	// ok
	{
		call := Call{
			ID:       dnTypes.NewZeroID(),
			UniqueID: "unique",
			Creator:  addr,
			Approved: true,
			Executed: true,
			Msg:      NewMockMsMsg("route", "type", true),
			MsgRoute: "route",
			MsgType:  "type",
			Height:   100,
		}
		require.NoError(t, call.Valid(200))
	}
}
