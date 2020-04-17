package yamltypes

type (
	Marshaler interface {
		MarshalYAML() (interface{}, error)
	}

	Unmarshaler interface {
		UnmarshalYAML(unmarshal func(interface{}) error) error
	}

	MarshalerUnmarshaler interface {
		Marshaler
		Unmarshaler
	}
)
