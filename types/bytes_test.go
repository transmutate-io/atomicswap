package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestBytes(t *testing.T) {
	b := Bytes("hello world")
	bb, err := yaml.Marshal(b)
	require.NoError(t, err, "can't marshal")
	b2 := Bytes{}
	err = yaml.Unmarshal(bb, &b2)
	require.NoError(t, err, "can't unmarshal")
	require.Equal(t, b, b2, "mismatch")
}
