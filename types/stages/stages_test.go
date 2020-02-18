package stages

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStages(t *testing.T) {
	for k, v := range _stages {
		require.Equal(t, k.String(), v, "name mismatch")
		require.NoError(t, k.Set(v), "can't set")
	}
	_, err := ParseStage("bad-stage")
	require.Error(t, err, "expecting an error")
	tests := []struct {
		sm  map[Stage]Stage
		exp []string
	}{
		{
			BuyerStages,
			[]string{
				"generate",
				"send-key",
				"receive-key",
				"receive-lock",
				"generate-lock",
				"send-lock",
				"wait-locked-funds",
				"lock-funds",
				"wait-redeem-funds",
				"redeem",
			},
		},
		{
			SellerStages,
			[]string{
				"generate",
				"receive-key",
				"send-key",
				"generate-lock",
				"send-lock",
				"receive-lock",
				"lock-funds",
				"wait-locked-funds",
				"redeem",
			},
		},
	}
	for _, tst := range tests {
		got := make([]string, 0, len(tst.exp))
		cs := Initialized
		for {
			ns := tst.sm[cs]
			if ns == Done {
				break
			}
			cs = ns
			got = append(got, ns.String())
		}
		require.Equal(t, tst.exp, got, "stage sequence mismatch")
	}
}
