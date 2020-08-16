package hash

type hasherBTC struct{}

// NewBTC returns an hasher for bitcoin
func NewBTC() Hasher { return hasherBTC{} }

// Hash256 implement Hasher
func (h hasherBTC) Hash256(b []byte) []byte { return Sha256Sum(Sha256Sum(b)) }

// Hash160 implement Hasher
func (h hasherBTC) Hash160(b []byte) []byte { return Ripemd160Sum(Sha256Sum(b)) }
