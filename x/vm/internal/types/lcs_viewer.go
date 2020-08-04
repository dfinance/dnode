package types

type ViewerType string

const (
	ViewerTypeU8      ViewerType = "U8"
	ViewerTypeU64     ViewerType = "U64"
	ViewerTypeU128    ViewerType = "U128"
	ViewerTypeBool    ViewerType = "bool"
	ViewerTypeAddress ViewerType = "address"
	ViewerTypeStruct  ViewerType = "struct"
	ViewerTypeVector  ViewerType = "vector"
)

type ViewerRequest []ViewerItem

type ViewerItem struct {
	Name      string         `json:"name"`
	Type      ViewerType     `json:"type"`
	InnerItem *ViewerRequest `json:"inner_item"`
}
