package bigint

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClamp(t *testing.T) {
	start := big.NewInt(1)
	end := big.NewInt(10)

	result := Clamp(start, end, 20)
	require.True(t, end == result)

	result = Clamp(start, end, 10)
	require.True(t, end == result)

	result = Clamp(start, end, 5)
	require.False(t, end == result)
	require.Equal(t, uint64(5), result.Uint64())
}
