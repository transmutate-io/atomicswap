package {{ .Values.package }}

type Private{{ .Values.short }} struct{ *PrivateBTC }

func ParsePrivate{{ .Values.short }}(b []byte) (Private, error) {
	return &Private{{ .Values.short }}{PrivateBTC: parsePrivateBTC(b)}, nil
}

func NewPrivate{{ .Values.short }}() (Private, error) {
	priv, err := newPrivateBTC()
	if err != nil {
		return nil, err
	}
	return &Private{{ .Values.short }}{PrivateBTC: priv}, nil
}

func (k *Private{{ .Values.short }}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	priv := &PrivateBTC{}
	if err := unmarshal(priv); err != nil {
		return err
	}
	k.PrivateBTC = priv
	return nil
}

type Public{{ .Values.short }} struct{ *PublicBTC }

func NewPublic{{ .Values.short }}(b []byte) (Public, error) {
	pub, err := newPublicBTC(b)
	if err != nil {
		return nil, err
	}
	return &Public{{ .Values.short }}{PublicBTC: pub}, nil
}

func (k *Public{{ .Values.short }}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	pub := &PublicBTC{}
	if err := unmarshal(pub); err != nil {
		return err
	}
	k.PublicBTC = pub
	return nil
}
