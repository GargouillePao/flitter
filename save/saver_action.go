package save

type item struct {
	Name   string
	Action SaveAction
	Value  interface{}
}

type SaveAction uint8

const (
	_ SaveAction = iota
	SA_Add
	SA_Remove
	SA_Clean
)
