package hash

type hasherDCR struct{}

// NewDCR returns an hasher for decred
func NewDCR() Hasher { return hasherDCR{} }

// Hash256 implement Hasher
func (h hasherDCR) Hash256(b []byte) []byte { return Blake256Sum(Blake256Sum(b)) }

// Hash160 implement Hasher
func (h hasherDCR) Hash160(b []byte) []byte { return Ripemd160Sum(Blake256Sum(b)) }
