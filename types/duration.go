package types

import (
	"time"
)

type Duration time.Duration

func (d Duration) MarshalYAML() (interface{}, error) { return time.Duration(d).String(), nil }

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
