package reflection

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReflections(t *testing.T) {
	a := &struct {
		A int64
		B float64
		C string
	}{}
	require.True(t, IsType(a.A, int64(42)), "type error")
	require.True(t, FieldIsType(a, "B", float64(3.14)), "type error")
	require.True(t, FieldIsType(*a, "C", ""), "type error")
	b, err := ReplaceTypeFields(a, FieldReplacementMap{
		"A": uint64(42),
		"B": float32(3.14),
	})
	require.NoError(t, err, "can't replace types")
	require.True(t, FieldIsType(b, "A", uint64(42)), "type error")
	require.True(t, FieldIsType(b, "B", float32(3.14)), "type error")
	require.True(t, FieldIsType(b, "C", ""), "type error")
	c, err := FilterFields(a, "A", "C")
	require.NoError(t, err, "can't filter")
	require.True(t, HasField(c, "A"), "A is missing")
	require.True(t, HasField(c, "C"), "C is missing")
	require.False(t, HasField(c, "B"), "B is present")
	d := &struct {
		A uint64
		B float32
	}{}
	err = CopyFieldsByName(c, d)
	require.NoError(t, err, "can't copy fields")
}
