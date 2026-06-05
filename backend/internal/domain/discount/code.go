package discount

import (
	"regexp"
	"strings"
)

var discountCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{2,63}$`)

// NormalizeCode uppercases and validates a manager-defined discount code.
func NormalizeCode(raw string) (string, error) {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if !discountCodePattern.MatchString(code) {
		return "", ErrInvalidCode
	}
	return code, nil
}
