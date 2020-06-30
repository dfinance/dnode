package types

// PathData is used to store VM paths in the storage (as JSON marshal used, we have to wrap it to struct).
type PathData struct {
	Path []byte `json:"path"`
}
