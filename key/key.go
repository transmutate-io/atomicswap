package key

import "transmutate.io/pkg/atomicswap/yamltypes"

type (
	NewPrivateFunc = func() (Private, error)
	NewPublicFunc  = func(b []byte) (Public, error)

	Keyer interface{ Key() interface{} }

	Private interface {
		Public() Public
		Sign(b []byte) ([]byte, error)
		yamltypes.MarshalerUnmarshaler
		Serialize() []byte
		Keyer
	}

	Public interface {
		Hash160() []byte
		Verify(sig, msg []byte) error
		yamltypes.MarshalerUnmarshaler
		SerializeCompressed() []byte
		Keyer
	}
)
