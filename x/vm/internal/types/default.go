package types

const (
	ModuleName   = "vm"
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	GovRouterKey = ModuleName
	//
	VmGasPrice       = 1 // gas unit price for VM execution
	VmUnknownTagType = -1
	// VM Event to sdk.Event conversion params
	EventTypeProcessingGas = 10000 // initial gas for processing event type.
	EventTypeNoGasLevels   = 2     // defines number of nesting levels that do not charge gas
)

var (
	KeyDelimiter   = []byte(":")
	KeyGenesisInit = []byte("gen") // is storage has that key, InitGenesis was done
)
