package usecase

import "strings"

// resolvePatchString applies PATCH semantics for optional nullable profile strings.
// Omitted field (nil pointer) keeps fallback; explicit empty string clears to nil.
func resolvePatchString(incoming, fallback *string) *string {
	if incoming == nil {
		return fallback
	}
	trimmed := strings.TrimSpace(*incoming)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
