package roles

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoles(t *testing.T) {
	for r, rn := range _roles {
		require.Equal(t, r.String(), rn, "name mismatch")
		require.NoError(t, r.Set(rn), "can't set")
	}
	_, err := ParseRole("bad-role")
	require.Error(t, err, "expecting an error")
}
