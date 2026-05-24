package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	lowerChars  = "abcdefghijkmnopqrstuvwxyz"
	upperChars  = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	digitChars  = "23456789"
	symbolChars = "!@#$%^&*"
)

// GenerateTempPassword returns a random password with complexity guarantees.
func GenerateTempPassword(length int) (string, error) {
	if length < 8 {
		return "", fmt.Errorf("temp password length must be at least 8")
	}
	all := lowerChars + upperChars + digitChars + symbolChars
	buf := make([]byte, length)

	required := make([]byte, 0, 4)
	for _, pool := range []string{lowerChars, upperChars, digitChars, symbolChars} {
		ch, err := randomChar(pool)
		if err != nil {
			return "", err
		}
		required = append(required, ch)
	}
	copy(buf, required)
	for i := len(required); i < length; i++ {
		ch, err := randomChar(all)
		if err != nil {
			return "", err
		}
		buf[i] = ch
	}
	if err := shuffleBytes(buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func randomChar(pool string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(pool))))
	if err != nil {
		return 0, fmt.Errorf("random int: %w", err)
	}
	return pool[n.Int64()], nil
}

func shuffleBytes(in []byte) error {
	for i := len(in) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return fmt.Errorf("shuffle index: %w", err)
		}
		in[i], in[j.Int64()] = in[j.Int64()], in[i]
	}
	return nil
}
