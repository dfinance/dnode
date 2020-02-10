// +build integ

package keeper

import (
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
	"wings-blockchain/x/vm/internal/types"
)

// Test deploy module with real VM.
func TestKeeper_DeployContract(t *testing.T) {
	input := setupTestInput(false)
	defer closeInput(input)

	// launch ds server
	rawServer := StartServer(input.vk.listener, input.vk.dsServer)
	defer rawServer.Stop()

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

// Test execute script with real VM.
func TestKeeper_ExecuteScript(t *testing.T) {
}