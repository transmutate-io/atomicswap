package types

import (
	"encoding/base64"
	"encoding/hex"

	"transmutate.io/pkg/atomicswap/hash"
)

// Bytes represents a yaml marshalable []byte
type Bytes []byte

func parse(s string, parser func(string) ([]byte, error)) (Bytes, error) {
	r, err := parser(s)
	if err != nil {
		return nil, err
	}
	return Bytes(r), nil
}

// ParseHex parses an hex string
func ParseHex(h string) (Bytes, error) { return parse(h, hex.DecodeString) }

// ParseBase64 parses a base64 string
func ParseBase64(b64 string) (Bytes, error) {
	return parse(b64, base64.RawStdEncoding.DecodeString)
}

// Hex returns an hex string
func (h Bytes) Hex() string { return hex.EncodeToString(h) }

// Base64 returns a base64 string
func (h Bytes) Base64() string { return base64.RawStdEncoding.EncodeToString(h) }

// Hash160 returns the ripemd160(sha256(b)) of the data
func (h Bytes) Hash160() Bytes { return hash.Hash160(h) }

// MarshalYAML implements yaml.Marshaler
func (h Bytes) MarshalYAML() (interface{}, error) { return h.Base64(), nil }

// UnmarshalYAML implements yaml.Unmarshaler
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
