package params

const (
	Bitcoin Crypto = iota
	Litecoin
	Ethereum
)

var _cryptoNames = map[Crypto]string{
	Bitcoin: "Bitcoin",
	Litecoin: "Litecoin",
	Ethereum: "Ethereum",
}

var _cryptoTickers = map[Crypto]string{
	Bitcoin: "BTC",
	Litecoin: "LTC",
	Ethereum: "ETH",
}

