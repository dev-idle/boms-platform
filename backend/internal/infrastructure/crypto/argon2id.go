// Package crypto provides cryptographic adapters (Argon2id password hashing).
package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"golang.org/x/crypto/argon2"
)

// Argon2Hasher implements port.PasswordHasher using Argon2id (PHC string format).
type Argon2Hasher struct {
	time       uint32
	memory     uint32
	threads    uint8
	saltLength uint32
	keyLength  uint32
}

// NewArgon2Hasher returns a hasher configured from application settings.
func NewArgon2Hasher(cfg config.Argon2Config) *Argon2Hasher {
	return &Argon2Hasher{
		time:       cfg.Iterations,
		memory:     cfg.Memory,
		threads:    cfg.Parallelism,
		saltLength: cfg.SaltLength,
		keyLength:  cfg.KeyLength,
	}
}

// Hash implements port.PasswordHasher.
func (h *Argon2Hasher) Hash(password string) (string, error) {
	if password == "" {
		return "", apperrors.ErrValidation.WithDetail("field", "password")
	}
	salt := make([]byte, h.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("argon2id salt: %w", err)
	}
	return h.encode(password, salt), nil
}

func (h *Argon2Hasher) encode(password string, salt []byte) string {
	hash := argon2.IDKey([]byte(password), salt, h.time, h.memory, h.threads, h.keyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.memory, h.time, h.threads, b64Salt, b64Hash)
}

// Verify implements port.PasswordHasher.
func (h *Argon2Hasher) Verify(encoded, password string) error {
	if password == "" {
		return apperrors.ErrInvalidCredentials
	}
	salt, expected, params, err := decodeHash(encoded)
	if err != nil {
		return apperrors.ErrInvalidCredentials
	}
	actual := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, params.keyLen)
	if subtle.ConstantTimeCompare(actual, expected) != 1 {
		return apperrors.ErrInvalidCredentials
	}
	return nil
}

// NeedsRehash implements port.PasswordHasher.
func (h *Argon2Hasher) NeedsRehash(encoded string) bool {
	_, _, params, err := decodeHash(encoded)
	if err != nil {
		return true
	}
	return params.time != h.time ||
		params.memory != h.memory ||
		params.threads != h.threads ||
		params.keyLen != h.keyLength
}

type hashParams struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

func decodeHash(encoded string) (salt, hash []byte, params hashParams, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return nil, nil, params, errors.New("invalid argon2id format")
	}
	var version int
	if _, err = fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return nil, nil, params, errors.New("unsupported argon2 version")
	}
	var time, memory uint32
	var threads uint8
	if _, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return nil, nil, params, errors.New("invalid argon2 params")
	}
	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, params, err
	}
	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, params, err
	}
	keyLen, err := safeUint32(len(hash))
	if err != nil {
		return nil, nil, params, err
	}
	params = hashParams{time: time, memory: memory, threads: threads, keyLen: keyLen}
	return salt, hash, params, nil
}

func safeUint32(n int) (uint32, error) {
	if n < 0 || n > int(^uint32(0)) {
		return 0, strconv.ErrRange
	}
	return uint32(n), nil
}

var _ port.PasswordHasher = (*Argon2Hasher)(nil)
