package keeper

import (
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
	"wings-blockchain/x/vm/internal/types"
)

// TODO: different mock servers for different responses?

// Deploy script with mocked VM.
func TestKeeper_DeployContractMock(t *testing.T) {
	input := setupTestInput()
	defer closeInput(input)

	acc := sdk.AccAddress(randomValue(20))

	codeBytes, err := hex.DecodeString(moveCode)
	if err != nil {
		t.Fatal(err)
	}

	msg := types.NewMsgDeployModule(acc, codeBytes)

	events, err := input.vk.DeployContract(input.ctx, msg)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, events, 1)
	require.EqualValues(t, types.EventTypeKeep, events[0].Type)
}

// Deploy script execute with mocked VM.
func TestKeeper_ExecuteScriptMock(t *testing.T) {
	input := setupTestInput()
	defer closeInput(input)

	acc := sdk.AccAddress(randomValue(20))

	codeBytes, err := hex.DecodeString(moveCode)
	if err != nil {
		t.Fatal(err)
	}

	msg := types.NewMsgExecuteScript(acc, codeBytes, nil)

	events, err := input.vk.ExecuteScript(input.ctx, msg)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, events, 2)
	require.EqualValues(t, types.EventTypeKeep, events[0].Type)
}
