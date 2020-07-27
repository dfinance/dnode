package v0_6

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/core/msmodule"
)

type (
	Params struct {
		IntervalToExecute int64 `json:"interval_to_execute"`
	}

	Call struct {
		ID       dnTypes.ID     `json:"id"`
		UniqueID string         `json:"unique_id"`
		Creator  sdk.AccAddress `json:"creator"`
		Approved bool           `json:"approved"`
		Executed bool           `json:"executed"`
		Failed   bool           `json:"failed"`
		Rejected bool           `json:"rejected"`
		Error    string         `json:"error"`
		Msg      msmodule.MsMsg `json:"msg_data"`
		MsgRoute string         `json:"msg_route"`
		MsgType  string         `json:"msg_type"`
		Height   int64          `json:"height"`
	}

	Votes []sdk.AccAddress

	GenesisCallItem struct {
		Call  Call  `json:"call"`
		Votes Votes `json:"votes"`
	}

	GenesisQueueItem struct {
		CallID      dnTypes.ID `json:"call_id"`
		BlockHeight int64      `json:"block_height"`
	}

	GenesisState struct {
		Parameters Params             `json:"parameters"`
		LastCallID *dnTypes.ID        `json:"last_call_id"`
		CallItems  []GenesisCallItem  `json:"call_items"`
		QueueItems []GenesisQueueItem `json:"queue_items"`
	}
)
