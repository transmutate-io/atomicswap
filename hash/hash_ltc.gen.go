package hash

type hasherLTC struct{ hasherBTC }

// NewLTC returns an hasher for litecoin
func NewLTC() Hasher { return hasherLTC{} }
