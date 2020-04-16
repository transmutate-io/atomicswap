package key

type (
	NewPrivateFunc = func() (Private, error)
	NewPublicFunc  = func(b []byte) (Public, error)

	Keyer interface{ Key() interface{} }

	YAMLMarshalerUnmarshaler interface {
		MarshalYAML() (interface{}, error)
		UnmarshalYAML(unmarshal func(interface{}) error) error
	}

	Private interface {
		Public() Public
		Sign(b []byte) ([]byte, error)
		YAMLMarshalerUnmarshaler
		Serialize() []byte
		Keyer
	}

	Public interface {
		Hash160() []byte
		Verify(sig, msg []byte) error
		YAMLMarshalerUnmarshaler
		SerializeCompressed() []byte
		Keyer
	}
)
