package catalog

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidSlug = errors.New("invalid catalog slug")
	slugPattern    = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

// NormalizeSlug lowercases and validates a URL-safe catalog slug.
func NormalizeSlug(slug string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(slug))
	if normalized == "" || !slugPattern.MatchString(normalized) {
		return "", ErrInvalidSlug
	}
	return normalized, nil
}
