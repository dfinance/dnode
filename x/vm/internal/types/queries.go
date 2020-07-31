package types

const (
	QueryValue   = "value"
	QueryLcsView = "lcsView"
)

// Client request for writeSet data.
type ValueReq struct {
	Address []byte `json:"address" yaml:"address"`
	Path    []byte `json:"path" yaml:"path"`
}

// Client request for LCS view writeSet data.
type LcsViewReq struct {
	Address     []byte        `json:"address" yaml:"address"`
	ModuleName  string        `json:"module_name" yaml:"module_name"`
	StructName  string        `json:"struct_name" yaml:"struct_name"`
	ViewRequest ViewerRequest `json:"view_request" yaml:"view_request"`
}

// Client response for writeSet data.
type ValueResp struct {
	Value string `json:"value" yaml:"value" format:"HEX string"`
}

func (resp ValueResp) String() string {
	return "Value: " + resp.Value
}
