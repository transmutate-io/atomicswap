package hash

type hasherDOGE struct{ hasherBTC }

// NewDOGE returns an hasher for dogecoin
func NewDOGE() Hasher { return hasherDOGE{} }
