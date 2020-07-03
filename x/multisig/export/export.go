// Export is used to prevent cycle dependency using /x/multisig/alias.go (by all multisig modules).
package export

import "github.com/dfinance/dnode/x/multisig/internal/types"

type (
	MsgSubmitCall    = types.MsgSubmitCall
	MsgConfirmCall   = types.MsgConfirmCall
	MsgRevokeConfirm = types.MsgRevokeConfirm
)

var (
	NewMsgSubmitCall    = types.NewMsgSubmitCall
	NewMsgConfirmCall   = types.NewMsgConfirmCall
	NewMsgRevokeConfirm = types.NewMsgRevokeConfirm
)
