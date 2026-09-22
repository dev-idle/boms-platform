package utils

import (
	"regexp"
	"strings"
)

// Vietnam mobile numbers — the only phones the bakery takes. The frontend
// applies the same rule (src/lib/validation/phone.ts); both are tested against
// contracts/vietnam-phone-cases.json.
var (
	// A carrier prefix MIC has assigned, then seven digits. 095 is retired; the
	// 11-digit 01x numbers moved to these prefixes in 2018. When MIC assigns a
	// new prefix, add it here, in the frontend and in the cases.
	mobileNumber = regexp.MustCompile(`^(?:3[2-9]|5[25689]|7[06-9]|8[1-9]|9[0-46-9])\d{7}$`)
	// Digits and the separators people type between them; anything else is not a phone.
	phoneCharacters = regexp.MustCompile(`^[\d\s.()+-]*$`)
	nonDigits       = regexp.MustCompile(`\D`)
)

// nationalNumber finds the national number inside whatever was typed. `00` is
// Vietnam's international access code and `84` the country code, so both come
// off the front of a long number; then the trunk `0` does, since no national
// number starts with 0. Zeros inside the number are never touched.
func nationalNumber(phone string) string {
	digits := strings.TrimPrefix(nonDigits.ReplaceAllString(phone, ""), "00")
	if strings.HasPrefix(digits, "84") && len(digits) > 9 {
		digits = digits[2:]
	}
	return strings.TrimLeft(digits, "0")
}

// NormalizeVietnamPhone returns the stored E.164 form (+84…) of a Vietnam mobile
// number and reports whether the input was one.
func NormalizeVietnamPhone(phone string) (string, bool) {
	if !phoneCharacters.MatchString(phone) {
		return "", false
	}
	national := nationalNumber(phone)
	if !mobileNumber.MatchString(national) {
		return "", false
	}
	return "+84" + national, true
}

// NormalizeVietnamPhonePtr normalizes an optional phone from a request. nil stays
// nil (absent) and a blank value becomes "" (a request to clear); anything else
// must be a Vietnam mobile number, and ok is false when it is not.
func NormalizeVietnamPhonePtr(phone *string) (normalized *string, ok bool) {
	if phone == nil {
		return nil, true
	}
	trimmed := strings.TrimSpace(*phone)
	if trimmed == "" {
		return &trimmed, true
	}
	e164, ok := NormalizeVietnamPhone(trimmed)
	if !ok {
		return nil, false
	}
	return &e164, true
}
