package key

// PrivateDOGE represents a private key for dogecoin
type PrivateDOGE struct{ *PrivateBTC }

// ParsePrivateDOGE parses a dogecoin private key
func ParsePrivateDOGE(b []byte) (Private, error) {
	return &PrivateDOGE{PrivateBTC: parsePrivateBTC(b)}, nil
}

// NewPrivateDOGE returns a new dogecoin private key
func NewPrivateDOGE() (Private, error) {
	priv, err := newPrivateBTC()
	if err != nil {
		return nil, err
	}
	return &PrivateDOGE{PrivateBTC: priv}, nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (k *PrivateDOGE) UnmarshalYAML(unmarshal func(interface{}) error) error {
	priv := &PrivateBTC{}
	if err := unmarshal(priv); err != nil {
		return err
	}
	k.PrivateBTC = priv
	return nil
}

// MarshalYAML implement yaml.Marshaler
func (k *PrivateDOGE) MarshalYAML() (interface{}, error) {
	if k == nil {
		return nil, nil
	}
	return k.PrivateBTC.MarshalYAML()
}

// PublicDOGE represents a dogecoin public key
type PublicDOGE struct{ *PublicBTC }

// ParsePublicDOGE parses a dogecoin public key
func ParsePublicDOGE(b []byte) (Public, error) {
	pub, err := parsePublicBTC(b)
	if err != nil {
		return nil, err
	}
	return &PublicDOGE{PublicBTC: pub}, nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (k *PublicDOGE) UnmarshalYAML(unmarshal func(interface{}) error) error {
	pub := &PublicBTC{}
	if err := unmarshal(pub); err != nil {
		return err
	}
	k.PublicBTC = pub
	return nil
}
