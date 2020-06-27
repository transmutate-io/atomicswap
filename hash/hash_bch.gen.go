package hash

type hasherBCH struct{ hasherBTC }

func NewBCH() Hasher { return hasherBCH{} }
