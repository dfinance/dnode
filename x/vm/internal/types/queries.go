package types

const (
	QueryValue = "value"
)

// Client request for writeSet data.
type ValueReq struct {
	Address []byte `json:"address" yaml:"address"`
	Path    []byte `json:"path" yaml:"path"`
}

// Client response for writeSet data.
type ValueResp struct {
	Value string `json:"value" yaml:"value" format:"HEX string"`
}

func (resp ValueResp) String() string {
	return "Value: " + resp.Value
}
