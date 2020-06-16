package types

type TableStringer interface {
	TableHeaders() []string
	TableValues() []string
}
