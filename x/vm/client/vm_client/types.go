package vm_client

const (
	CodeTypeModule = "module"
	CodeTypeScript = "script"
)

// CompiledItems struct contains multiple CompiledItem.
type CompiledItems []CompiledItem

// CompiledItem struct contains code from file and meta information.
type CompiledItem struct {
	Code     string         `json:"code"`
	ByteCode []byte         `json:"-"`
	Name     string         `json:"name"`
	Methods  []ModuleMethod `json:"methods,omitempty"`
	Types    []ModuleType   `json:"types,omitempty"`
	CodeType string         `json:"code_type"`
}

type ModuleType struct {
	Name           string            `json:"name"`
	IsResource     bool              `json:"resource"`
	TypeParameters []string          `json:"type_parameters"`
	Field          []ModuleTypeField `json:"properties"`
}

type ModuleTypeField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ModuleMethod struct {
	Name           string   `json:"name"`
	IsPublic       bool     `json:"public"`
	IsNative       bool     `json:"native"`
	TypeParameters []string `json:"type_parameters"`
	Arguments      []string `json:"arguments"`
	Returns        []string `json:"returns"`
}
