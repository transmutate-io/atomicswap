package roles

import "fmt"

type InvalidRoleError string

func (e InvalidRoleError) Error() string {
	return fmt.Sprintf("invalid role: \"%s\"", string(e))
}

type Role int

func ParseRole(s string) (Role, error) {
	var r Role
	if err := (&r).Set(s); err != nil {
		return 0, err
	}
	return r, nil
}

func (v Role) String() string { return _Role[v] }

func (v *Role) Set(sv string) error {
	nv, ok := _RoleNames[sv]
	if !ok {
		return InvalidRoleError(sv)
	}
	*v = nv
	return nil
}

func (v Role) MarshalYAML() (interface{}, error) { return v.String(), nil }

func (v *Role) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var r string
	if err := unmarshal(&r); err != nil {
		return err
	}
	return v.Set(r)
}

const (
	Buyer Role = iota
	Seller
)

var (
	_Role = map[Role]string{
		Buyer: "buyer",
		Seller: "seller",
	}
	_RoleNames map[string]Role
)

func init() {
	_RoleNames = make(map[string]Role, len(_Role))
	for k, v := range _Role {
		_RoleNames[v] = k
	}
}

