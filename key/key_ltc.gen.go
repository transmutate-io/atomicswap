package key

// PrivateLTC represents a private key for litecoin
type PrivateLTC struct{ *PrivateBTC }

// ParsePrivateLTC parses a litecoin private key
func ParsePrivateLTC(b []byte) (Private, error) {
	return &PrivateLTC{PrivateBTC: parsePrivateBTC(b)}, nil
}

// NewPrivateLTC returns a new litecoin private key
func NewPrivateLTC() (Private, error) {
	priv, err := newPrivateBTC()
	if err != nil {
		return nil, err
	}
	return &PrivateLTC{PrivateBTC: priv}, nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (k *PrivateLTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	priv := &PrivateBTC{}
	if err := unmarshal(priv); err != nil {
		return err
	}
	k.PrivateBTC = priv
	return nil
}

// MarshalYAML implement yaml.Marshaler
func (k *PrivateLTC) MarshalYAML() (interface{}, error) {
	if k == nil {
		return nil, nil
	}
	return k.PrivateBTC.MarshalYAML()
}

// PublicLTC represents a litecoin public key
type PublicLTC struct{ *PublicBTC }

// ParsePublicLTC parses a litecoin public key
func ParsePublicLTC(b []byte) (Public, error) {
	pub, err := parsePublicBTC(b)
	if err != nil {
		return nil, err
	}
	return &PublicLTC{PublicBTC: pub}, nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (k *PublicLTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	pub := &PublicBTC{}
	if err := unmarshal(pub); err != nil {
		return err
	}
	k.PublicBTC = pub
	return nil
}
