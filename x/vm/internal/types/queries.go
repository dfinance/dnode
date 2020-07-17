package types

// Query when access path to read value.
type QueryAccessPath struct {
	Address []byte `json:"address" yaml:"address"`
	Path    []byte `json:"path" yaml:"path"`
}

// Query response.
type QueryValueResp struct {
	Value string `json:"value" yaml:"value" format:"HEX string"`
}

func (resp QueryValueResp) String() string {
	return "Value: " + resp.Value
}
