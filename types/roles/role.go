package roles

import "fmt"

type InvalidRoleError string

func (e InvalidRoleError) Error() string { return fmt.Sprintf("invalid role: \"%s\"", string(e)) }

type Role int

const (
	Buyer Role = iota
	Seller
)

var (
	_roles      = map[Role]string{Buyer: "buyer", Seller: "seller"}
	_rolesNames = map[string]Role{"buyer": Buyer, "seller": Seller}
)

func ParseRole(role string) (Role, error) {
	var r Role
	if err := (&r).Set(role); err != nil {
		return 0, err
	}
	return r, nil
}

func (r Role) String() string { return _roles[r] }

func (r *Role) Set(role string) error {
	ns, ok := _rolesNames[role]
	if !ok {
		return InvalidRoleError(role)
	}
	*r = ns
	return nil
}
