package types

import (
	"time"
)

// Duration represents a yaml marshalable time.Duration
type Duration time.Duration

// MarshalYAML implements yaml.Marshaler
func (d Duration) MarshalYAML() (interface{}, error) { return time.Duration(d).String(), nil }

// UnmarshalYAML implements yaml.Unmarshaler
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	nd, err := time.ParseDuration(r)
	if err != nil {
		return err
	}
	*d = Duration(nd)
	return nil
}
