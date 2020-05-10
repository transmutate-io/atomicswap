package tx

var txFuncs = map[string]NewTxFunc{
	"bitcoin-cash": NewBCH,
	"bitcoin":      NewBTC,
	"dogecoin":     NewDOGE,
	"litecoin":     NewLTC,
}
