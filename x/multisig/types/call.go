package types

// Call that will be executed itself, contains msg instances, that executing via router and hadler
type Call struct {
	Approved bool
	Executed bool

	Msg MsMsg
}

// Create new call instance
func NewCall(msg MsMsg) Call {
	return Call{
		Approved: false,
		Executed: false,
		Msg: msg,
	}
}