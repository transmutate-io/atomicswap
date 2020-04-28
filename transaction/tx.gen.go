package transaction

var txFuncs = map[string]NewTxFunc{
	"bitcoin-cash": NewTxBCH,
	"bitcoin":      NewTxBTC,
	"dogecoin":     NewTxDOGE,
	"litecoin":     NewTxLTC,
}
