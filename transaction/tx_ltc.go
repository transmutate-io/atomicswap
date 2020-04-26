package transaction

type txLTC struct{ *txBTC }

// NewTxLTC creates a new *txLTC
func NewTxLTC() (Tx, error) {
	b, _ := NewTxBTC()
	return &txLTC{txBTC: b.(*txBTC)}, nil
}

func (tx *txLTC) Copy() Tx { return &txLTC{txBTC: tx.txBTC.Copy().(*txBTC)} }

func (tx *txLTC) MarshalYAML() (interface{}, error) { return tx.txBTC, nil }

func (tx *txLTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(tx.txBTC)
}
