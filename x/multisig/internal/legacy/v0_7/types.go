package v0_7

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/core/msmodule"
	v06 "github.com/dfinance/dnode/x/multisig/internal/legacy/v0_6"
)

type (
	Call struct {
		ID       dnTypes.ID     `json:"id"`
		UniqueID string         `json:"unique_id"`
		Creator  sdk.AccAddress `json:"creator"`
		Approved bool           `json:"approved"`
		Executed bool           `json:"executed"`
		Rejected bool           `json:"rejected"`
		Error    string         `json:"error"`
		Msg      msmodule.MsMsg `json:"msg_data"`
		MsgRoute string         `json:"msg_route"`
		MsgType  string         `json:"msg_type"`
		Height   int64          `json:"height"`
	}

	GenesisCallItem struct {
		Call  Call      `json:"call"`
		Votes v06.Votes `json:"votes"`
	}

	GenesisState struct {
		Parameters v06.Params             `json:"parameters"`
		LastCallID *dnTypes.ID            `json:"last_call_id"`
		CallItems  []GenesisCallItem      `json:"call_items"`
		QueueItems []v06.GenesisQueueItem `json:"queue_items"`
	}
)
