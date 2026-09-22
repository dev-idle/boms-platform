package utils

import (
	"regexp"
	"strings"
)

// Vietnam numbers in the spellings people type: national (leading zero) and
// international (+84 or 84, sometimes followed by the national zero). After the
// prefix comes the national number, checked against the MIC numbering plan:
//
//   - mobile, 9 digits: an assigned carrier prefix — 32–39, 52 55 56 58 59, 70
//     76–79, 81–89, 90–94, 96–99 — then 7 digits. 095 is retired, and the 11-digit
//     01x numbers moved to these prefixes on 2018-09-15.
//   - land line, 10 digits: 2, a province code (24x Hanoi, 28x Ho Chi Minh City,
//     one per province elsewhere), then a 7-digit subscriber number.
//
// Service numbers (1800, 1900, 11x) and 069 are not personal phones and fail.
// The frontend applies the same pattern (src/lib/validation/phone.ts); both are
// tested against contracts/vietnam-phone-cases.json. When MIC assigns a new
// prefix, add it in both places and a case to that file.
var (
	vietnamPhonePattern = regexp.MustCompile(
		`^(?:0|\+?840?)(` +
			`(?:3[2-9]|5[25689]|7[06-9]|8[1-9]|9[0-46-9])\d{7}` +
			`|2(?:0[3-9]|1[0-689]|2[0-25-9]|3[2-9]|4[2-8]|5[124-9]|6[0-39]|7[0-7]|8[2-7]|9[0-4679])\d{7}` +
			`)$`,
	)
	phoneSeparatorsRegex = regexp.MustCompile(`[\s.()\-]`)
)

// NormalizeVietnamPhone returns the stored E.164 form (+84…) of a Vietnam phone
// number and reports whether the input was one. Separators are how people type,
// not what is stored, so they are dropped before matching.
func NormalizeVietnamPhone(phone string) (string, bool) {
	m := vietnamPhonePattern.FindStringSubmatch(phoneSeparatorsRegex.ReplaceAllString(phone, ""))
	if m == nil {
		return "", false
	}
	return "+84" + m[1], true
}

// NormalizeVietnamPhonePtr normalizes an optional phone from a request. nil stays
// nil (absent) and a blank value becomes "" (a request to clear); anything else
// must be a Vietnam number, and ok is false when it is not.
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
