package jwt_test

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"
	"time"

	"github.com/boms/backend/internal/config"
	jwtinfra "github.com/boms/backend/internal/infrastructure/jwt"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSigner(t *testing.T) *jwtinfra.EdDSASigner {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	cfg := config.JWTConfig{
		Ed25519PrivateKey: base64.StdEncoding.EncodeToString(seed),
		Issuer:            "boms-api",
		Audience:          "boms",
		AccessTTL:         15 * time.Minute,
		RefreshTTL:        7 * 24 * time.Hour,
	}
	s, err := jwtinfra.NewEdDSASigner(cfg)
	require.NoError(t, err)
	return s
}

func TestEdDSASigner(t *testing.T) {
	t.Parallel()
	s := testSigner(t)

	t.Run("access round trip", func(t *testing.T) {
		t.Parallel()
		raw, err := s.SignAccess(port.AccessTokenClaims{
			Subject: "user-1", Role: "customer", SessionID: "sid-1", JTI: "jti-a",
		})
		require.NoError(t, err)
		claims, err := s.ParseAccess(raw)
		require.NoError(t, err)
		assert.Equal(t, "user-1", claims.Subject)
		assert.Equal(t, "customer", claims.Role)
	})

	t.Run("refresh round trip without role", func(t *testing.T) {
		t.Parallel()
		raw, err := s.SignRefresh(port.RefreshTokenClaims{
			Subject: "user-1", SessionID: "sid-1", JTI: "jti-r",
		})
		require.NoError(t, err)
		claims, err := s.ParseRefresh(raw)
		require.NoError(t, err)
		assert.Equal(t, "jti-r", claims.JTI)
	})

	t.Run("cross use reject", func(t *testing.T) {
		t.Parallel()
		access, err := s.SignAccess(port.AccessTokenClaims{Subject: "u", Role: "customer", SessionID: "s", JTI: "j"})
		require.NoError(t, err)
		_, err = s.ParseRefresh(access)
		require.Error(t, err)
	})

	t.Run("HS256 reject", func(t *testing.T) {
		t.Parallel()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user", "token_use": "access", "iss": "boms-api", "aud": "boms",
		})
		raw, err := token.SignedString([]byte("super-secret-key-at-least-32-bytes!!"))
		require.NoError(t, err)
		_, err = s.ParseAccess(raw)
		require.Error(t, err)
	})

	t.Run("none alg reject", func(t *testing.T) {
		t.Parallel()
		_, err := s.ParseAccess("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1In0.")
		require.Error(t, err)
	})

	t.Run("expired returns apperrors.ErrTokenExpired", func(t *testing.T) {
		t.Parallel()
		seed, err := base64.StdEncoding.DecodeString(mustSeed())
		require.NoError(t, err)
		priv := ed25519.NewKeyFromSeed(seed)
		claims := jwt.MapClaims{
			"iss": "boms-api", "aud": "boms", "sub": "u", "token_use": "access",
			"role": "customer", "sid": "s", "jti": "j",
			"exp": time.Now().Add(-time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
		raw, err := token.SignedString(priv)
		require.NoError(t, err)
		_, parseErr := s.ParseAccess(raw)
		require.Error(t, parseErr)
		assert.ErrorIs(t, parseErr, apperrors.ErrTokenExpired)
	})

	t.Run("tampered signature reject", func(t *testing.T) {
		t.Parallel()
		raw, err := s.SignAccess(port.AccessTokenClaims{Subject: "u", Role: "customer", SessionID: "s", JTI: "j"})
		require.NoError(t, err)
		require.Greater(t, len(raw), 10)
		tampered := raw[:len(raw)-2] + "xx"
		_, err = s.ParseAccess(tampered)
		require.Error(t, err)
		_, ok := apperrors.AsAppError(err)
		assert.True(t, ok)
	})

	t.Run("missing sid or jti reject", func(t *testing.T) {
		t.Parallel()
		seed, _ := base64.StdEncoding.DecodeString(mustSeed())
		priv := ed25519.NewKeyFromSeed(seed)
		claims := jwt.MapClaims{
			"iss": "boms-api", "aud": "boms", "sub": "u", "token_use": "access",
			"role": "customer",
			"exp":  time.Now().Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
		raw, err := token.SignedString(priv)
		require.NoError(t, err)
		_, err = s.ParseAccess(raw)
		require.Error(t, err)
	})
}

func mustSeed() string {
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	return base64.StdEncoding.EncodeToString(seed)
}
