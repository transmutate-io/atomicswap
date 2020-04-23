package key

type PrivateDOGE struct{ *PrivateBTC }

func ParsePrivateDOGE(b []byte) (Private, error) {
	return &PrivateDOGE{PrivateBTC: parsePrivateBTC(b)}, nil
}

func NewPrivateDOGE() (Private, error) {
	priv, err := newPrivateBTC()
	if err != nil {
		return nil, err
	}
	return &PrivateDOGE{PrivateBTC: priv}, nil
}

func (k *PrivateDOGE) UnmarshalYAML(unmarshal func(interface{}) error) error {
	priv := &PrivateBTC{}
	if err := unmarshal(priv); err != nil {
		return err
	}
	k.PrivateBTC = priv
	return nil
}

type PublicDOGE struct{ *PublicBTC }

func NewPublicDOGE(b []byte) (Public, error) {
	pub, err := newPublicBTC(b)
	if err != nil {
		return nil, err
	}
	return &PublicDOGE{PublicBTC: pub}, nil
}

func (k *PublicDOGE) UnmarshalYAML(unmarshal func(interface{}) error) error {
	pub := &PublicBTC{}
	if err := unmarshal(pub); err != nil {
		return err
	}
	k.PublicBTC = pub
	return nil
}
