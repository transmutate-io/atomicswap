package key

import "transmutate.io/pkg/atomicswap/types"

type (
	NewPrivateFunc = func() (Private, error)
	NewPublicFunc  = func(b []byte) (Public, error)

	Keyer interface{ Key() interface{} }

	Private interface {
		Public() Public
		Sign(b []byte) ([]byte, error)
		types.YAMLMarshalerUnmarshaler
		Serialize() []byte
		Keyer
	}

	Public interface {
		Hash160() []byte
		Verify(sig, msg []byte) error
		types.YAMLMarshalerUnmarshaler
		SerializeCompressed() []byte
		Keyer
	}
)
