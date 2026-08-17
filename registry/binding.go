package registry

type visibility uint8

const (
	loaded visibility = iota
	registered
	exported
)

type binding struct {
	key         string
	visibility  visibility
	declaration *declaration
}
