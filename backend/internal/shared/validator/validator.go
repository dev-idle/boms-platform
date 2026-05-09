package validator

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	v     *validator.Validate
	vOnce sync.Once
)

// V returns a shared *validator.Validate instance (lazy init).
func V() *validator.Validate {
	vOnce.Do(func() {
		v = validator.New()
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
