package key

type (
	NewPrivateFunc = func() (Private, error)
	NewPublicFunc  = func(b []byte) (Public, error)

	Keyer interface{ Key() interface{} }

	Private interface {
		MarshalYAML() (interface{}, error)
		UnmarshalYAML(unmarshal func(interface{}) error) error
		Public() Public
		Sign(b []byte) ([]byte, error)
		Keyer
	}

	Public interface {
		MarshalYAML() (interface{}, error)
		UnmarshalYAML(unmarshal func(interface{}) error) error
		Hash160() []byte
		Verify(sig, msg []byte) error
		Keyer
	}
)
