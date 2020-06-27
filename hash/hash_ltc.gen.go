package hash

type hasherLTC struct{ hasherBTC }

func NewLTC() Hasher { return hasherLTC{} }
