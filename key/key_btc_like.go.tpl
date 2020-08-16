package {{ .Values.package }}

// Private{{ .Values.short }} represents a private key for {{ .Values.name }}
type Private{{ .Values.short }} struct{ *PrivateBTC }

// ParsePrivate{{ .Values.short }} parses a {{ .Values.name }} private key
func ParsePrivate{{ .Values.short }}(b []byte) (Private, error) {
	return &Private{{ .Values.short }}{PrivateBTC: parsePrivateBTC(b)}, nil
}

// NewPrivate{{ .Values.short }} returns a new {{ .Values.name }} private key
func NewPrivate{{ .Values.short }}() (Private, error) {
	priv, err := newPrivateBTC()
	if err != nil {
		return nil, err
	}
	return &Private{{ .Values.short }}{PrivateBTC: priv}, nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (k *Private{{ .Values.short }}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	priv := &PrivateBTC{}
	if err := unmarshal(priv); err != nil {
		return err
	}
	k.PrivateBTC = priv
	return nil
}

// MarshalYAML implement yaml.Marshaler
func (k *Private{{ .Values.short }}) MarshalYAML() (interface{}, error) {
	if k == nil {
		return nil, nil
	}
	return k.PrivateBTC.MarshalYAML()
}

// Public{{ .Values.short }} represents a {{ .Values.name }} public key
type Public{{ .Values.short }} struct{ *PublicBTC }

// ParsePublic{{ .Values.short }} parses a {{ .Values.name }} public key
func ParsePublic{{ .Values.short }}(b []byte) (Public, error) {
	pub, err := parsePublicBTC(b)
	if err != nil {
		return nil, err
	}
	return &Public{{ .Values.short }}{PublicBTC: pub}, nil
}

// UnmarshalYAML implement yaml.Unmarshaler
func (k *Public{{ .Values.short }}) UnmarshalYAML(unmarshal func(interface{}) error) error {
	pub := &PublicBTC{}
	if err := unmarshal(pub); err != nil {
		return err
	}
	k.PublicBTC = pub
	return nil
}
