package transaction

type txDOGE struct{ *txBTC }

// NewTxDOGE creates a new *txDOGE
func NewTxDOGE() (Tx, error) {
	b, _ := NewTxBTC()
	return &txDOGE{txBTC: b.(*txBTC)}, nil
}

func (tx *txDOGE) Copy() Tx { return &txDOGE{txBTC: tx.txBTC.Copy().(*txBTC)} }

func (tx *txDOGE) MarshalYAML() (interface{}, error) { return tx.txBTC, nil }

func (tx *txDOGE) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(tx.txBTC)
}
