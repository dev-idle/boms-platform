package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeSeedPhone(t *testing.T) {
	t.Parallel()

	t.Run("blank_seeds_no_phone", func(t *testing.T) {
		t.Parallel()
		got, err := normalizeSeedPhone("   ")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("valid_number_is_stored_normalized", func(t *testing.T) {
		t.Parallel()
		got, err := normalizeSeedPhone("0912 345 678")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "+84912345678", *got)
	})

	t.Run("invalid_number_stops_the_seed", func(t *testing.T) {
		t.Parallel()
		got, err := normalizeSeedPhone("555-0100")
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}
