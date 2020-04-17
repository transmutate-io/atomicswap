package transaction

import "fmt"

type InvalidTransactionTypeError string

func (e InvalidTransactionTypeError) Error() string { return fmt.Sprintf("invalid transactiontype: \"%s\"", string(e)) }

type TransactionType int

func ParseTransactionType(s string) (TransactionType, error) {
	var r TransactionType
	if err := (&r).Set(s); err != nil {
		return 0, err
	}
	return r, nil
}

func (v TransactionType) String() string { return _TransactionType[v] }

func (v *TransactionType) Set(sv string) error {
	nv, ok := _TransactionTypeNames[sv]
	if !ok {
		return InvalidTransactionTypeError(sv)
	}
	*v = nv
	return nil
}

func (v TransactionType) MarshalYAML() (interface{}, error) { return v.String(), nil }

func (v *TransactionType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	return v.Set(r)
}

const (
 	UTXO TransactionType = iota
 	StateBased
)

var (
	_TransactionType = map[TransactionType]string{
		UTXO:       "utxo",
		StateBased: "state-based",
	}
	_TransactionTypeNames map[string]TransactionType
)

func init() {
	_TransactionTypeNames = make(map[string]TransactionType, len(_TransactionType))
	for k, v := range _TransactionType {
		_TransactionTypeNames[v] = k
	}
}
