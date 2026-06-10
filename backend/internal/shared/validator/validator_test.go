package validator_test

import (
	"testing"

	"github.com/boms/backend/internal/shared/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type passwordField struct {
	Password string `validate:"password_complexity"`
}

func TestPasswordComplexity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "lowercase letter and digit", password: "rose2024", wantErr: false},
		{name: "uppercase letter and digit", password: "Rose2024", wantErr: false},
		{name: "unicode letter and digit", password: "café9", wantErr: false},
		{name: "letters only", password: "rosebuds", wantErr: true},
		{name: "digits only", password: "12345678", wantErr: true},
		{name: "symbols only", password: "!@#$%^&*", wantErr: true},
		{name: "letter without digit", password: "rosebud!", wantErr: true},
		{name: "digit without letter", password: "1234567!", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validator.Struct(passwordField{Password: tt.password})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
