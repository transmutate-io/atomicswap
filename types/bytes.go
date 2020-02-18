package types

import (
	"encoding/base64"
	"encoding/hex"
)

type Bytes []byte

func parse(s string, parser func(string) ([]byte, error)) (Bytes, error) {
	r, err := parser(s)
	if err != nil {
		return nil, err
	}
	return Bytes(r), nil
}

func ParseHex(h string) (Bytes, error) { return parse(h, hex.DecodeString) }

func ParseBase64(b64 string) (Bytes, error) {
	return parse(b64, base64.RawStdEncoding.DecodeString)
}

func (h Bytes) Hex() string    { return hex.EncodeToString(h) }
func (h Bytes) Base64() string { return base64.RawStdEncoding.EncodeToString(h) }

func (h Bytes) MarshalYAML() (interface{}, error) { return h.Base64(), nil }

func (h *Bytes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	b, err := ParseBase64(r)
	if err != nil {
		return err
	}
	*h = b
	return nil
}
