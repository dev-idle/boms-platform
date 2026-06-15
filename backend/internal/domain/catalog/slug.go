package catalog

import (
	"errors"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	ErrInvalidSlug = errors.New("invalid catalog slug")
	slugPattern    = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	slugSeparator  = regexp.MustCompile(`[^a-z0-9]+`)
	stripDiacritics = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)))
)

// NormalizeSlug lowercases and validates a URL-safe catalog slug.
func NormalizeSlug(slug string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(slug))
	if normalized == "" || !slugPattern.MatchString(normalized) {
		return "", ErrInvalidSlug
	}
	return normalized, nil
}

// SlugFromName derives a URL-safe slug from a display name (create fallback).
// Diacritics are stripped to match frontend slugifyCatalogName (NFD fold).
func SlugFromName(name string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(name))
	folded, _, err := transform.String(stripDiacritics, normalized)
	if err != nil {
		return "", ErrInvalidSlug
	}
	normalized = slugSeparator.ReplaceAllString(folded, "-")
	normalized = strings.Trim(normalized, "-")
	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}
	if len(normalized) > 128 {
		normalized = strings.Trim(normalized[:128], "-")
	}
	return NormalizeSlug(normalized)
}
