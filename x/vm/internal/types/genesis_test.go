// +build unit

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVM_Genesis_Validate(t *testing.T) {
	t.Parallel()

	const (
		address1Ok      = "010203040506070809A0A1A2A3A4A5A6A7A8A9AB"
		address2Ok      = "010203040506070809A0A1A2A3A4A5A6A7A8A9AC"
		address3Ok      = "010203040506070809A0A1A2A3A4A5A6A7A8A9AD"
		addressWrongLen = "010203040506070809A0A1A2A3A4A5A6A7A8A9"
		path1Ok         = "B0B1B2B3B4B5B6B7B8C0"
		path2Ok         = "B0B1B2B3B4B5B6B7B8C1"
		path3Ok         = "B0B1B2B3B4B5B6B7B8C2"
		valueOk         = "112233445566778899AABBCCDDEEFF"
		wrongHex        = "xxyyzz"
	)

	// fail: address wrong HEX
	{
		state := GenesisState{
			WriteSet: []GenesisWriteOp{
				{
					Address: wrongHex,
					Path:    path1Ok,
					Value:   valueOk,
				},
			},
		}
		require.Error(t, state.Validate())
	}

	// fail: address wrong length
	{
		state := GenesisState{
			WriteSet: []GenesisWriteOp{
				{
					Address: addressWrongLen,
					Path:    path1Ok,
					Value:   valueOk,
				},
			},
		}
		require.Error(t, state.Validate())
	}

	// fail: path wrong HEX
	{
		state := GenesisState{
			WriteSet: []GenesisWriteOp{
				{
					Address: address1Ok,
					Path:    wrongHex,
					Value:   valueOk,
				},
			},
		}
		require.Error(t, state.Validate())
	}

	// fail: value wrong HEX
	{
		state := GenesisState{
			WriteSet: []GenesisWriteOp{
				{
					Address: address1Ok,
					Path:    path1Ok,
					Value:   wrongHex,
				},
			},
		}
		require.Error(t, state.Validate())
	}

	// fail: duplicated writeOp
	{
		state := GenesisState{
			WriteSet: []GenesisWriteOp{
				{
					Address: address1Ok,
					Path:    path1Ok,
					Value:   valueOk,
				},
				{
					Address: address2Ok,
					Path:    path2Ok,
					Value:   valueOk,
				},
				{
					Address: address1Ok,
					Path:    path1Ok,
					Value:   valueOk,
				},
			},
		}
		require.Error(t, state.Validate())
	}

	// ok
	{
		state := GenesisState{
			WriteSet: []GenesisWriteOp{
				{
					Address: address1Ok,
					Path:    path1Ok,
					Value:   valueOk,
				},
				{
					Address: address2Ok,
					Path:    path2Ok,
					Value:   valueOk,
				},
				{
					Address: address3Ok,
					Path:    path3Ok,
					Value:   valueOk,
				},
			},
		}
		require.NoError(t, state.Validate())
	}
}
