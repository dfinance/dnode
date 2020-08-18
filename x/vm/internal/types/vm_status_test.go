// +build unit

package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
)

// Test NewVMStatus.
func TestVM_NewVMStatus(t *testing.T) {
	status := "error"
	message := "out of gas"

	vmStatus := NewVMStatus(status, message)
	require.Equal(t, vmStatus.Status, status)
	require.Equal(t, vmStatus.Message, message)

	vmStatus = NewVMStatus(AttributeValueStatusKeep, "")
	require.Empty(t, vmStatus.Message)
}

// Test NewTxVMResponse.
func TestVM_NewTxVMStatus(t *testing.T) {
	statuses := make(VMStatuses, 3)

	statuses[0] = NewVMStatus("error", "")
	statuses[1] = NewVMStatus("discard", "invalid signature")
	statuses[2] = NewVMStatus("error", "bad u32")

	txHash := "00"
	txVMStatus := NewTxVMStatus(txHash, statuses)

	require.Equal(t, txHash, txVMStatus.Hash)
	require.EqualValues(t, txVMStatus.VMStatuses, statuses)
}

// New NewVMStatusFromABCILogs.
func TestVM_NewVMStatusFromABCILogs(t *testing.T) {
	msgs := make([]string, 2)
	msgs[0] = "out of gas"
	msgs[1] = "bad u32"

	hash := "01"
	txResp := types.TxResponse{
		TxHash: hash,
		Logs: types.ABCIMessageLogs{
			types.NewABCIMessageLog(0, "",
				NewContractEvents(&vm_grpc.VMExecuteResponse{
					Status: &vm_grpc.VMStatus{
						Error: &vm_grpc.VMStatus_Abort{
							Abort: &vm_grpc.Abort{},
						},
						Message: &vm_grpc.Message{
							Text: msgs[0],
						},
					},
				}),
			),
			types.NewABCIMessageLog(1, "",
				NewContractEvents(&vm_grpc.VMExecuteResponse{
					Status: &vm_grpc.VMStatus{
						Error: &vm_grpc.VMStatus_Abort{
							Abort: &vm_grpc.Abort{},
						},
						Message: &vm_grpc.Message{
							Text: msgs[1],
						},
					},
				}),
			),
			types.NewABCIMessageLog(2, "",
				NewContractEvents(&vm_grpc.VMExecuteResponse{
					Status: &vm_grpc.VMStatus{},
				}),
			),
		},
	}

	statuses := NewVMStatusFromABCILogs(txResp)
	require.Equal(t, hash, statuses.Hash)
	require.Len(t, statuses.VMStatuses, len(txResp.Logs))

	for i, st := range statuses.VMStatuses {
		for _, ev := range txResp.Logs[i].Events {
			for _, atr := range ev.Attributes {
				if atr.Key == "message" {
					require.Equal(t, st.Message, atr.Value)
				}
				if atr.Key == "status" {
					require.Equal(t, st.Status, atr.Value)
				}
			}
		}
	}
}
