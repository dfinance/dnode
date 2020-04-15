package types

// Query when access path to read value.
type QueryAccessPath struct {
	Address []byte `json:"address"`
	Path    []byte `json:"path"`
}

// Query response.
type QueryValueResp struct {
	Value string `json:"value" format:"HEX string"`
}

func (resp QueryValueResp) String() string {
	return "Value: " + resp.Value
}
