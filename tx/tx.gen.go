package tx

var txFuncs = map[string]NewTxFunc{
	"bitcoin-cash": NewBCH,
	"bitcoin":      NewBTC,
	"decred":       NewDCR,
	"dogecoin":     NewDOGE,
	"litecoin":     NewLTC,
}
