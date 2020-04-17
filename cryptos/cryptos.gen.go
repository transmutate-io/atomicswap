package cryptos

var (
	Cryptos = map[string]newCryptoFunc{
		"bitcoin": newCryptoBTC,
		"litecoin": newCryptoLTC,
		"dogecoin": newCryptoDOGE,
		"bitcoin-cash": newCryptoBCH,
		}
)
