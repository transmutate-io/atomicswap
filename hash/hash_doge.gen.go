package hash

type hasherDOGE struct{ hasherBTC }

func NewDOGE() Hasher { return hasherDOGE{} }
