package stages

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestStager(t *testing.T) {
	s := NewStager(LockFunds, RedeemFunds, Done)
	b, err := yaml.Marshal(s)
	require.NoError(t, err, "can't marshal")
	s2 := NewStager()
	err = yaml.Unmarshal(b, s2)
	require.NoError(t, err, "can't unmarshal")
	require.Equal(t, s.stages, s2.stages, "mismatch")
	require.Equal(t, s.NextStage(), s2.NextStage(), "next stage mismatch")
	require.Equal(t, s.stages, s2.stages, "mismatch")
}
