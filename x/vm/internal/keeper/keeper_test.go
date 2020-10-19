// +build unit

package keeper

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

// Deploy script with mocked VM.
func TestVMKeeper_DeployContractMock(t *testing.T) {
	t.Parallel()

	input := newTestInput(true)
	defer input.Stop()

	acc := sdk.AccAddress(randomValue(20))

	codeBytes, err := hex.DecodeString(moveCode)
	if err != nil {
		t.Fatal(err)
	}

	msg := types.NewMsgDeployModule(acc, []types.Contract{codeBytes})

	err = input.vk.DeployContract(input.ctx, msg)
	if err != nil {
		t.Fatal(err)
	}

	events := input.ctx.EventManager().Events()

	require.Len(t, events, 2)

	require.EqualValues(t, sdk.EventTypeMessage, events[0].Type)
	require.EqualValues(t, sdk.AttributeKeyModule, events[0].Attributes[0].Key)
	require.EqualValues(t, types.ModuleName, events[0].Attributes[0].Value)

	require.EqualValues(t, types.EventTypeContractStatus, events[1].Type)
	require.EqualValues(t, types.AttributeStatus, events[1].Attributes[0].Key)
	require.EqualValues(t, types.AttributeValueStatusKeep, events[1].Attributes[0].Value)
}

// Deploy script execute with mocked VM.
func TestVMKeeper_ExecuteScriptMock(t *testing.T) {
	t.Parallel()

	input := newTestInput(true)
	defer input.Stop()

	acc := sdk.AccAddress(randomValue(20))

	codeBytes, err := hex.DecodeString(moveCode)
	if err != nil {
		t.Fatal(err)
	}

	msg := types.NewMsgExecuteScript(acc, codeBytes, nil)

	err = input.vk.ExecuteScript(input.ctx, msg)
	if err != nil {
		t.Fatal(err)
	}

	events := input.ctx.EventManager().Events()

	require.Len(t, events, 3)

	require.EqualValues(t, sdk.EventTypeMessage, events[0].Type)
	require.EqualValues(t, sdk.AttributeKeyModule, events[0].Attributes[0].Key)
	require.EqualValues(t, types.ModuleName, events[0].Attributes[0].Value)

	require.EqualValues(t, types.EventTypeContractStatus, events[1].Type)
	require.EqualValues(t, types.AttributeStatus, events[1].Attributes[0].Key)
	require.EqualValues(t, types.AttributeValueStatusKeep, events[1].Attributes[0].Value)
}
