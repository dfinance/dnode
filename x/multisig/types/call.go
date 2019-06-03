package types

type Call struct {
	Approved bool
	Executed bool

	Msg MsMsg
}

func NewCall(msg MsMsg) Call {
	return Call{
		Approved: false,
		Executed: false,
		Msg: msg,
	}
}