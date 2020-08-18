package types

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
)

// VMStatus is a VM error response.
type VMStatus struct {
	Status  string `json:"status"`            // Status of error: error/discard
	Message string `json:"message,omitempty"` // Message
}

func (status VMStatus) String() string {
	return fmt.Sprintf("VM status:\n"+
		"  Status: %s\n"+
		"  Message:  %s",
		status.Status, status.Message,
	)
}

// NewVMStatus creates a new VMStatus error.
func NewVMStatus(status, message string) VMStatus {
	return VMStatus{
		Status:  status,
		Message: message,
	}
}

// Slice of VMStatus objects (VM error responses).
type VMStatuses []VMStatus

func (list VMStatuses) String() string {
	strBuilder := strings.Builder{}
	strBuilder.WriteString("VMStatuses:\n")
	for i, status := range list {
		strBuilder.WriteString(status.String())
		if i < len(list)-1 {
			strBuilder.WriteString("\n")
		}
	}

	return strBuilder.String()
}

// TxVMStatus is a response containing TX hash with VM errors.
type TxVMStatus struct {
	Hash       string     `json:"hash"`
	VMStatuses VMStatuses `json:"vm_status"`
}

func (tx TxVMStatus) String() string {
	return fmt.Sprintf("Tx:\n"+
		"  Hash: %s\n"+
		"  Statuses: %s",
		tx.Hash, tx.VMStatuses.String(),
	)
}

// NewTxVMStatus creates a new TxVMStatus object.
func NewTxVMStatus(hash string, statuses VMStatuses) TxVMStatus {
	return TxVMStatus{
		Hash:       hash,
		VMStatuses: statuses,
	}
}

// NewVMStatusFromABCILogs converts SDK TxResponse log events to TxVMStatus.
func NewVMStatusFromABCILogs(tx types.TxResponse) TxVMStatus {
	statuses := make(VMStatuses, 0)

	for _, log := range tx.Logs {
		for _, event := range log.Events {
			isFound := false

			if event.Type == EventTypeContractStatus {
				status := ""
				message := ""

				for _, attr := range event.Attributes {
					// find that it's event contains contract status.
					if attr.Key == AttributeStatus {
						status = attr.Value

						if status == AttributeValueStatusDiscard {
							isFound = true
							break
						}
					}
				}

				// event found.
				if isFound {
					for _, attr := range event.Attributes {
						if attr.Key == AttributeErrMessage {
							message = attr.Value
						}
					}
				}

				statuses = append(statuses, NewVMStatus(status, message))
			}
		}
	}

	return NewTxVMStatus(tx.TxHash, statuses)
}
