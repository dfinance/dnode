package types

// Call that will be executed itself, contains msg instances, that executing via router and hadler
type Call struct {
	// When call approved to execute
	Approved bool

	// Execution failed or executed
	Executed bool
	Failed   bool

	// If call was rejected
	Rejected bool
	Error    string

	// Msg to execute
	Msg MsMsg

	// Height when call submitted
	height int64
}

// Create new call instance
func NewCall(msg MsMsg, height int64) Call {
	return Call{
		Approved: false,
		Executed: false,
		Rejected: false,
		Failed:   false,
		Msg: 	  msg,
		height:   height,
	}
}

func (call Call) GetHeight() int64 {
	return call.height
}