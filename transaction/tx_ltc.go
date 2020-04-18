package transaction

type txLTC struct{ *txBTC }

// NewLTC creates a new *txLTC
func NewLTC() Tx { return &txLTC{txBTC: NewBTC().(*txBTC)} }

func (tx *txLTC) Copy() Tx { return &txLTC{txBTC: tx.txBTC.Copy().(*txBTC)} }

func (tx *txLTC) MarshalYAML() (interface{}, error) { return tx.txBTC, nil }

func (tx *txLTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(tx.txBTC)
}
