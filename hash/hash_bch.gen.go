package hash

type hasherBCH struct{ hasherBTC }

// NewBCH returns an hasher for bitcoin-cash
func NewBCH() Hasher { return hasherBCH{} }
