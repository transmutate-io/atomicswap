package transaction

type txDOGE struct{ *txBTC }

// NewDOGE creates a new *txDOGE
func NewDOGE() Tx { return &txDOGE{txBTC: NewBTC().(*txBTC)} }

func (tx *txDOGE) Copy() Tx { return &txDOGE{txBTC: tx.txBTC.Copy().(*txBTC)} }

func (tx *txDOGE) MarshalYAML() (interface{}, error) { return tx.txBTC, nil }

func (tx *txDOGE) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(tx.txBTC)
}
