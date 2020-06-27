package hash

type hasherBTC struct{}

func NewBTC() Hasher { return hasherBTC{} }

func (h hasherBTC) Hash256(b []byte) []byte { return Sha256Sum(Sha256Sum(b)) }
func (h hasherBTC) Hash160(b []byte) []byte { return Ripemd160Sum(Sha256Sum(b)) }
