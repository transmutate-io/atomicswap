package types

type (
	YAMLMarshaler interface {
		MarshalYAML() (interface{}, error)
	}

	YAMLUnmarshaler interface {
		UnmarshalYAML(unmarshal func(interface{}) error) error
	}

	YAMLMarshalerUnmarshaler interface {
		YAMLMarshaler
		YAMLUnmarshaler
	}
)
