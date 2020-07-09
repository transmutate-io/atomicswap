package key

type PrivateLTC struct{ *PrivateBTC }

func ParsePrivateLTC(b []byte) (Private, error) {
	return &PrivateLTC{PrivateBTC: parsePrivateBTC(b)}, nil
}

func NewPrivateLTC() (Private, error) {
	priv, err := newPrivateBTC()
	if err != nil {
		return nil, err
	}
	return &PrivateLTC{PrivateBTC: priv}, nil
}

func (k *PrivateLTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	priv := &PrivateBTC{}
	if err := unmarshal(priv); err != nil {
		return err
	}
	k.PrivateBTC = priv
	return nil
}

func (k *PrivateLTC) MarshalYAML() (interface{}, error) {
	if k == nil {
		return nil, nil
	}
	return k.PrivateBTC.MarshalYAML()
}

type PublicLTC struct{ *PublicBTC }

func NewPublicLTC(b []byte) (Public, error) {
	pub, err := newPublicBTC(b)
	if err != nil {
		return nil, err
	}
	return &PublicLTC{PublicBTC: pub}, nil
}

func (k *PublicLTC) UnmarshalYAML(unmarshal func(interface{}) error) error {
	pub := &PublicBTC{}
	if err := unmarshal(pub); err != nil {
		return err
	}
	k.PublicBTC = pub
	return nil
}
