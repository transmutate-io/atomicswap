package hash

var newHasherFuncs = map[string]func() Hasher{
	"bitcoin-cash": NewBCH,
	"bitcoin":      NewBTC,
	"decred":       NewDCR,
	"dogecoin":     NewDOGE,
	"litecoin":     NewLTC,
}
