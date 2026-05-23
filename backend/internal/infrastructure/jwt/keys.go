package jwt

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/boms/backend/internal/config"
)

// EnsureDevSeed fills JWT Ed25519 seed in development when unset.
func EnsureDevSeed(cfg *config.JWTConfig, devMode bool) error {
	if cfg == nil {
		return errors.New("jwt config is nil")
	}
	if cfg.Ed25519PrivateKey != "" {
		return nil
	}
	if !devMode {
		return errors.New("jwt.ed25519_private_key is required outside development")
	}
	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		return err
	}
	cfg.Ed25519PrivateKey = base64.StdEncoding.EncodeToString(seed)
	return nil
}
