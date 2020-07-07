package rest

import (
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

//nolint:deadcode,unused
type (
	OrdersRespGetOrders struct {
		Height int64        `json:"height"`
		Result types.Orders `json:"result"`
	}

	OrdersRespGetOrder struct {
		Height int64       `json:"height"`
		Result types.Order `json:"result"`
	}

	OrdersRespRevokeOrder struct {
		Type  string `json:"type" yaml:"type"`
		Value struct {
			Msg        RevokeOrderMsg           `json:"msg" yaml:"msg"`
			Fee        authTypes.StdFee         `json:"fee" yaml:"fee"`
			Signatures []authTypes.StdSignature `json:"signatures" yaml:"signatures"`
			Memo       string                   `json:"memo" yaml:"memo"`
		} `json:"value" yaml:"type"`
	}

	RevokeOrderMsg struct {
		Type  string               `json:"type" yaml:"type"`
		Value types.MsgRevokeOrder `json:"value" yaml:"type"`
	}

	OrdersRespPostOrder struct {
		Type  string `json:"type" yaml:"type"`
		Value struct {
			Msg        PostOrderMsg             `json:"msg" yaml:"msg"`
			Fee        authTypes.StdFee         `json:"fee" yaml:"fee"`
			Signatures []authTypes.StdSignature `json:"signatures" yaml:"signatures"`
			Memo       string                   `json:"memo" yaml:"memo"`
		} `json:"value" yaml:"type"`
	}

	PostOrderMsg struct {
		Type  string             `json:"type" yaml:"type"`
		Value types.MsgPostOrder `json:"value" yaml:"type"`
	}
)
