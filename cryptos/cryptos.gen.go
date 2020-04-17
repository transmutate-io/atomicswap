package cryptos

var (
	_cryptos = map[string]newCryptoFunc{
		"bitcoin": newCryptoBTC,
		"litecoin": newCryptoLTC,
		"dogecoin": newCryptoDOGE,
		"bitcoin-cash": newCryptoBCH,
		}
)
