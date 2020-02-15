package params

var (
	_cryptosByName   map[string]Crypto
	_cryptosByTicker map[string]Crypto
)

func init() {
	_cryptosByName = make(map[string]Crypto, len(_cryptoNames))
	for c, n := range _cryptoNames {
		_cryptosByName[n] = c
	}
	_cryptosByTicker = make(map[string]Crypto, len(_cryptoNames))
	for c, n := range _cryptoTickers {
		_cryptosByTicker[n] = c
	}
}
