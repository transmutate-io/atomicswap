package hash

type hasherDCR struct{}

func NewDCR() Hasher { return hasherDCR{} }

func (h hasherDCR) Hash256(b []byte) []byte { return Blake256Sum(Blake256Sum(b)) }
func (h hasherDCR) Hash160(b []byte) []byte { return Ripemd160Sum(Blake256Sum(b)) }
