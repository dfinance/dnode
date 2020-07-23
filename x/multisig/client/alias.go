package client

import "github.com/dfinance/dnode/x/multisig/internal/types"

type (
	MsgSubmitCall    = types.MsgSubmitCall
	MsgConfirmCall   = types.MsgConfirmCall
	MsgRevokeConfirm = types.MsgRevokeConfirm
)

const (
	// Permissions
	PermRead  = types.PermRead
	PermWrite = types.PermWrite
)

var (
	RegisterMultiSigTypeCodec = types.RegisterMultiSigTypeCodec
	//
	NewMsgSubmitCall    = types.NewMsgSubmitCall
	NewMsgConfirmCall   = types.NewMsgConfirmCall
	NewMsgRevokeConfirm = types.NewMsgRevokeConfirm
)
