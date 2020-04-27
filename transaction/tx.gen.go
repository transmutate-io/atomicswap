package transaction

var txFuncs = map[string]NewTxFunc{
	"bitcoin": NewTxBTC,
	"litecoin": NewTxLTC,
	"dogecoin": NewTxDOGE,
	"bitcoin-cash": NewTxBCH,
}
