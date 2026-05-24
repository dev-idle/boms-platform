package validator

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	v     *validator.Validate
	vOnce sync.Once
)

// V returns a shared *validator.Validate instance (lazy init with custom rules).
func V() *validator.Validate {
	vOnce.Do(func() {
		v = validator.New()
		if err := v.RegisterValidation("password_complexity", passwordComplexity); err != nil {
			panic(fmt.Sprintf("register password_complexity: %v", err))
		}
	})
	return v
}

// Struct validates s using struct tags; returns a human-readable error when invalid.
func Struct(s any) error {
	if err := V().Struct(s); err != nil {
		return formatValidationError(err)
	}
	return nil
}

// FieldErrors maps validation failures to JSON field keys.
func FieldErrors(err error) map[string]string {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return nil
	}
	out := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		key := strings.ToLower(fe.Field())
		out[key] = fe.Tag()
	}
	return out
}

func passwordComplexity(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	var upper, digit bool
	for _, r := range s {
		if unicode.IsUpper(r) {
			upper = true
		}
		if unicode.IsDigit(r) {
			digit = true
		}
	}
	return upper && digit
}

func formatValidationError(err error) error {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return err
	}
	var b strings.Builder
	for i, fe := range verrs {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(fe.Field())
		b.WriteString(": ")
		b.WriteString(fe.Tag())
	}
	return fmt.Errorf("validation: %s", b.String())
}
