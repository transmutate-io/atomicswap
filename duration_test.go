package atomicswap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestDuration(t *testing.T) {
	b := Duration(25*time.Hour + time.Minute + time.Second)
	bb, err := yaml.Marshal(b)
	require.NoError(t, err, "can't marshal")
	b2 := Duration(0)
	err = yaml.Unmarshal(bb, &b2)
	require.NoError(t, err, "can't unmarshal")
	require.Equal(t, b, b2, "mismatch")
}
