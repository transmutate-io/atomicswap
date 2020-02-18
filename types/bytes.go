package types

import (
	"encoding/base64"
	"encoding/hex"
)

type Bytes []byte

func (h Bytes) Hex() string    { return hex.EncodeToString(h) }
func (h Bytes) Base64() string { return base64.RawStdEncoding.EncodeToString(h) }

func (h Bytes) MarshalYAML() (interface{}, error) { return h.Base64(), nil }

func (h *Bytes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	b, err := base64.RawStdEncoding.DecodeString(r)
	if err != nil {
		return err
	}
	*h = b
	return nil
}
