package utils

import "strings"

// NormalizeEmail lowercases and trims an email address for consistent lookup and storage.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
