// +build unit

package types

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
)

const (
	ERR_GAS      = "4002"
	ERR_GAS_MSG  = "OUT_OF_GAS"
	ERR_ZERO_SUB = "0"

	ERR_SIG = "1"
	ERR_U32 = "3018"

	ERR_U32_INT = 3018
	ERR_GAS_INT = 4002
)

// Test NewVMStatus.
func TestVM_NewVMStatus(t *testing.T) {
	status := "error"
	message := "out of gas"

	vmStatus := NewVMStatus(status, ERR_GAS, ERR_ZERO_SUB, message)
	require.Equal(t, vmStatus.Status, status)
	require.Equal(t, vmStatus.MajorCode, ERR_GAS)
	require.Equal(t, vmStatus.SubCode, ERR_ZERO_SUB)
	require.Equal(t, vmStatus.Message, message)
	require.Equal(t, ERR_GAS_MSG, vmStatus.StrCode)

	vmStatus = NewVMStatus(AttributeValueStatusKeep, "", "", "")
	require.Empty(t, vmStatus.StrCode)
	require.Empty(t, vmStatus.Message)
	require.Empty(t, vmStatus.MajorCode)
	require.Empty(t, vmStatus.SubCode)
}

// Test NewTxVMResponse.
func TestVM_NewTxVMStatus(t *testing.T) {
	statuses := make(VMStatuses, 3)

	statuses[0] = NewVMStatus("error", ERR_GAS, "0", "")
	statuses[1] = NewVMStatus("discard", ERR_SIG, "0", "invalid signature")
	statuses[2] = NewVMStatus("error", ERR_U32, "0", "bad u32")

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

	codes := make([]uint64, 2)
	codes[0] = ERR_GAS_INT
	codes[1] = ERR_U32_INT

	strCodes := make([]string, 2)
	strCodes[0] = strconv.FormatUint(codes[0], 10)
	strCodes[1] = strconv.FormatUint(codes[1], 10)

	hash := "01"
	txResp := types.TxResponse{
		TxHash: hash,
		Logs: types.ABCIMessageLogs{
			types.NewABCIMessageLog(0, "",
				NewContractEvents(&vm_grpc.VMExecuteResponse{
					Status: vm_grpc.ContractStatus_Keep,
					StatusStruct: &vm_grpc.VMStatus{
						MajorStatus: codes[0],
						SubStatus:   0,
						Message:     msgs[0],
					},
				}),
			),
			types.NewABCIMessageLog(1, "",
				NewContractEvents(&vm_grpc.VMExecuteResponse{
					Status: vm_grpc.ContractStatus_Discard,
					StatusStruct: &vm_grpc.VMStatus{
						MajorStatus: codes[1],
						SubStatus:   0,
						Message:     msgs[1],
					},
				}),
			),
			types.NewABCIMessageLog(2, "",
				NewContractEvents(&vm_grpc.VMExecuteResponse{Status: vm_grpc.ContractStatus_Keep}),
			),
		},
	}

	status := NewVMStatusFromABCILogs(txResp)
	require.Equal(t, hash, status.Hash)
	require.Len(t, status.VMStatuses, len(txResp.Logs))

	for i, code := range strCodes {
		isFound := false

		for _, status := range status.VMStatuses {
			if status.MajorCode == code {
				require.False(t, isFound)

				require.Equal(t, msgs[i], status.Message)
				require.Equal(t, ERR_ZERO_SUB, status.SubCode)

				isFound = true
			}
		}

		require.True(t, isFound, fmt.Sprintf("not found code %s", code))
	}
}
