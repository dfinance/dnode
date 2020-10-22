// +build unit

package vm_client

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/types_grpc"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/dfinance/dnode/x/common_vm"
)

func Test_NewAddressScriptArg(t *testing.T) {
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// ok
	{
		tag, err := NewAddressScriptArg(addr.String())
		require.NoError(t, err)
		require.Equal(t, types_grpc.VMTypeTag_Address, tag.Type)
		require.Equal(t, common_vm.Bech32ToLibra(addr), tag.Value)
	}

	// empty
	{
		_, err := NewAddressScriptArg("")
		require.Error(t, err)
	}

	// invalid
	{
		_, err := NewAddressScriptArg("invalid")
		require.Error(t, err)
	}
}

func Test_NewU8ScriptArg(t *testing.T) {
	// ok
	{
		tag, err := NewU8ScriptArg("128")
		require.NoError(t, err)
		require.Equal(t, types_grpc.VMTypeTag_U8, tag.Type)
		require.Equal(t, []byte{0x80}, tag.Value)
	}

	// empty
	{
		_, err := NewU8ScriptArg("")
		require.Error(t, err)
	}

	// invalid
	{
		_, err := NewU8ScriptArg("abc")
		require.Error(t, err)
	}

	// invalid: bitLen
	{
		_, err := NewU8ScriptArg("1000")
		require.Error(t, err)
	}
}

func Test_NewU64ScriptArg(t *testing.T) {
	// ok
	{
		tag, err := NewU64ScriptArg("305441741")
		require.NoError(t, err)
		require.Equal(t, types_grpc.VMTypeTag_U64, tag.Type)
		require.Equal(t, []byte{0xCD, 0xAB, 0x34, 0x12, 0x00, 0x00, 0x00, 0x00}, tag.Value)
	}

	// empty
	{
		_, err := NewU64ScriptArg("")
		require.Error(t, err)
	}

	// invalid
	{
		_, err := NewU64ScriptArg("abc")
		require.Error(t, err)
	}

	// invalid: bitLen
	{
		_, err := NewU64ScriptArg("100000000000000000000")
		require.Error(t, err)
	}
}

func Test_NewU128ScriptArg(t *testing.T) {
	// ok
	{
		tag, err := NewU128ScriptArg("1339673755198158349044581307228491775")
		require.NoError(t, err)
		require.Equal(t, types_grpc.VMTypeTag_U128, tag.Type)
		require.Equal(t, []byte{0xFF, 0xF, 0xE, 0xD, 0xC, 0xB, 0xA, 0x9, 0x8, 0x7, 0x6, 0x5, 0x4, 0x3, 0x2, 0x1}, tag.Value)
	}

	// ok: extending with zeros
	{
		tag, err := NewU128ScriptArg("18591708106338011145")
		require.NoError(t, err)
		require.Equal(t, types_grpc.VMTypeTag_U128, tag.Type)
		require.Equal(t, []byte{0x9, 0x8, 0x7, 0x6, 0x5, 0x4, 0x3, 0x2, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, tag.Value)
	}

	// empty
	{
		_, err := NewU128ScriptArg("")
		require.Error(t, err)
	}

	// invalid
	{
		_, err := NewU128ScriptArg("abc")
		require.Error(t, err)
	}

	// invalid: bitLen
	{
		_, err := NewU128ScriptArg("87112285931760246646623899502532662132735")
		require.Error(t, err)
	}
}

func Test_NewBoolScriptArg(t *testing.T) {
	// ok: true
	{
		{
			tag, err := NewBoolScriptArg("true")
			require.NoError(t, err)
			require.Equal(t, types_grpc.VMTypeTag_Bool, tag.Type)
			require.Equal(t, []byte{1}, tag.Value)
		}
		{
			tag, err := NewBoolScriptArg("True")
			require.NoError(t, err)
			require.Equal(t, types_grpc.VMTypeTag_Bool, tag.Type)
			require.Equal(t, []byte{1}, tag.Value)
		}
		{
			tag, err := NewBoolScriptArg("TRUE")
			require.NoError(t, err)
			require.Equal(t, types_grpc.VMTypeTag_Bool, tag.Type)
			require.Equal(t, []byte{1}, tag.Value)
		}
	}

	// ok: false
	{
		{
			tag, err := NewBoolScriptArg("false")
			require.NoError(t, err)
			require.Equal(t, types_grpc.VMTypeTag_Bool, tag.Type)
			require.Equal(t, []byte{0}, tag.Value)
		}
		{
			tag, err := NewBoolScriptArg("False")
			require.NoError(t, err)
			require.Equal(t, types_grpc.VMTypeTag_Bool, tag.Type)
			require.Equal(t, []byte{0}, tag.Value)
		}
		{
			tag, err := NewBoolScriptArg("FALSE")
			require.NoError(t, err)
			require.Equal(t, types_grpc.VMTypeTag_Bool, tag.Type)
			require.Equal(t, []byte{0}, tag.Value)
		}
	}

	// empty
	{
		_, err := NewBoolScriptArg("")
		require.Error(t, err)
	}

	// invalid
	{
		_, err := NewBoolScriptArg("abc")
		require.Error(t, err)
	}
}

func Test_NewVectorScriptArg(t *testing.T) {
	// ok
	{
		tag, err := NewVectorScriptArg("01020304")
		require.NoError(t, err)
		require.Equal(t, types_grpc.VMTypeTag_Vector, tag.Type)
		require.Equal(t, []byte{0x1, 0x2, 0x3, 0x4}, tag.Value)
	}

	// ok: prefixed
	{
		tag, err := NewVectorScriptArg("0xFFFEFD")
		require.NoError(t, err)
		require.Equal(t, types_grpc.VMTypeTag_Vector, tag.Type)
		require.Equal(t, []byte{0xFF, 0xFE, 0xFD}, tag.Value)
	}

	// empty
	{
		_, err := NewVectorScriptArg("")
		require.Error(t, err)
	}

	// invalid
	{
		_, err := NewVectorScriptArg("zzxxcc")
		require.Error(t, err)
	}
}
