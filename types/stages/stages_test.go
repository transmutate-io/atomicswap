package stages

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStages(t *testing.T) {
	_, err := ParseStage("invalid-stage")
	require.Error(t, err, "expecting and error")
	for k, v := range stages {
		s, err := ParseStage(v)
		require.NoError(t, err, "can't parse")
		require.Equal(t, k, s, "mismatch")
	}
}
