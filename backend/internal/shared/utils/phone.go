package utils

import (
	"regexp"
	"strings"
)

// Vietnam numbers in the spellings people type: national (leading zero) and
// international (+84 or 84, sometimes followed by the national zero). After the
// prefix comes either a mobile number — 3, 5, 7, 8 or 9 then eight digits — or a
// land line: 2, the rest of the area code and the subscriber number, ten digits
// in all. Nothing has been assigned to 0, 1, 4 or 6 since the 2018 renumber.
// The frontend applies the same pattern (src/lib/validation/phone.ts); both are
// tested against contracts/vietnam-phone-cases.json.
var (
	vietnamPhonePattern  = regexp.MustCompile(`^(?:0|\+?840?)(2\d{9}|[35789]\d{8})$`)
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
