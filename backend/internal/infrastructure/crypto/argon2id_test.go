package crypto_test

import (
	"testing"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/infrastructure/crypto"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultArgon2Config() config.Argon2Config {
	return config.Argon2Config{
		Memory:      65536,
		Iterations:  3,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}
}

func TestArgon2Hasher(t *testing.T) {
	t.Parallel()

	h := crypto.NewArgon2Hasher(defaultArgon2Config())

	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "round trip",
			fn: func(t *testing.T) {
				hash, err := h.Hash("correct horse battery staple")
				require.NoError(t, err)
				assert.Contains(t, hash, ",p=1$")
				require.NoError(t, h.Verify(hash, "correct horse battery staple"))
			},
		},
		{
			name: "wrong password",
			fn: func(t *testing.T) {
				hash, err := h.Hash("secret")
				require.NoError(t, err)
				err = h.Verify(hash, "wrong")
				require.Error(t, err)
				ae, ok := apperrors.AsAppError(err)
				require.True(t, ok)
				assert.Equal(t, apperrors.ErrInvalidCredentials.Code, ae.Code)
			},
		},
		{
			name: "salt randomness",
			fn: func(t *testing.T) {
				h1, err := h.Hash("same-password")
				require.NoError(t, err)
				h2, err := h.Hash("same-password")
				require.NoError(t, err)
				assert.NotEqual(t, h1, h2)
			},
		},
		{
			name: "malformed PHC",
			fn: func(t *testing.T) {
				err := h.Verify("$argon2id$invalid", "x")
				require.Error(t, err)
			},
		},
		{
			name: "needs rehash when params change",
			fn: func(t *testing.T) {
				old := crypto.NewArgon2Hasher(config.Argon2Config{
					Memory: 32 * 1024, Iterations: 3, Parallelism: 4, SaltLength: 16, KeyLength: 32,
				})
				hash, err := old.Hash("pw")
				require.NoError(t, err)
				assert.True(t, h.NeedsRehash(hash))
				fresh, err := h.Hash("pw")
				require.NoError(t, err)
				assert.False(t, h.NeedsRehash(fresh))
			},
		},
		{
			name: "empty password rejected",
			fn: func(t *testing.T) {
				_, err := h.Hash("")
				require.Error(t, err)
				ae, ok := apperrors.AsAppError(err)
				require.True(t, ok)
				assert.Equal(t, apperrors.ErrValidation.Code, ae.Code)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.fn(t)
		})
	}
}
