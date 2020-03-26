package keeper

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

// Deploy script with mocked VM.
func TestKeeper_DeployContractMock(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

	acc := sdk.AccAddress(randomValue(20))

	codeBytes, err := hex.DecodeString(moveCode)
	if err != nil {
		t.Fatal(err)
	}

	msg := types.NewMsgDeployModule(acc, codeBytes)

	err = input.vk.DeployContract(input.ctx, msg)
	if err != nil {
		t.Fatal(err)
	}

	events := input.ctx.EventManager().Events()

	require.Len(t, events, 1)
	require.EqualValues(t, types.EventTypeContractStatus, events[0].Type)
	require.EqualValues(t, types.AttrKeyStatus, events[0].Attributes[0].Key)
	require.EqualValues(t, types.StatusKeep, events[0].Attributes[0].Value)
}

// Deploy script execute with mocked VM.
func TestKeeper_ExecuteScriptMock(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

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

	require.Len(t, events, 2)
	require.EqualValues(t, types.EventTypeContractStatus, events[0].Type)
	require.EqualValues(t, types.AttrKeyStatus, events[0].Attributes[0].Key)
	require.EqualValues(t, types.StatusKeep, events[0].Attributes[0].Value)
}

// Check genesis Import / Export functionality
func TestKeeper_ExportGenesis(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

	// check export with no initial genesis
	{
		outputState := input.vk.ExportGenesis(input.ctx)
		require.Empty(t, outputState.WriteSet)
	}

	// initial state
	inputState := types.GenesisState{
		WriteSet: []types.GenesisWriteOp{
			{
				Address: "616464726573735f31", // address_1
				Path:    "706174685f31",       // path_1
				Value:   "76616c75655f31",     // value_1
			},
			{
				Address: "616464726573735f32", // address_2
				Path:    "706174685f32",       // path_2
				Value:   "76616c75655f32",     // value_2
			},
		},
	}

	// add non-init WriteSets
	input.vk.SetValue(input.ctx, &vm_grpc.VMAccessPath{
		Address: []byte("616464726573735f33"), // address_3
		Path:    []byte("706174685f33"),       // path_3
	}, []byte("76616c75655f33")) // value_3

	// check export with initial genesis
	{
		input.vk.InitGenesis(input.ctx, input.cdc.MustMarshalJSON(inputState))
		outputState := input.vk.ExportGenesis(input.ctx)
		require.Equal(t, inputState, outputState)
	}
}
