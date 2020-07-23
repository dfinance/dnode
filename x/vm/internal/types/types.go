package types

// Contract defines Move contract data (bytes).
type Contract []byte

// PathData is used to store VM paths in the storage (as JSON marshal used, we have to wrap it to struct).
type PathData struct {
	Path []byte `json:"path"`
}
