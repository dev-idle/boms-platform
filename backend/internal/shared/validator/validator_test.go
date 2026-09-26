package validator_test

import (
	"strings"
	"testing"

	"github.com/boms/backend/internal/dto"
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

// Mirrors the request DTOs: optional, and only checked when present.
type phoneField struct {
	Phone *string `json:"phone" validate:"omitempty,vn_phone"`
}

func TestVietnamPhone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		phone   *string
		wantErr bool
	}{
		{name: "absent", phone: nil, wantErr: false},
		{name: "blank clears", phone: strPtr(""), wantErr: false},
		{name: "whitespace clears", phone: strPtr("   "), wantErr: false},
		{name: "national mobile", phone: strPtr("0912 345 678"), wantErr: false},
		{name: "international mobile", phone: strPtr("+84912345678"), wantErr: false},
		{name: "land line", phone: strPtr("028 3822 1234"), wantErr: true},
		{name: "symbols", phone: strPtr("!!!!!!"), wantErr: true},
		{name: "too short", phone: strPtr("0912"), wantErr: true},
		{name: "foreign", phone: strPtr("+1 415 555 0172"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validator.Struct(phoneField{Phone: tt.phone})
			if !tt.wantErr {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), "vn_phone")
		})
	}
}

func strPtr(value string) *string {
	return &value
}

// The frontend maps details onto form fields by their JSON names, so the keys
// must be "full_name", not the Go field "FullName".
type profileRequest struct {
	FullName string  `json:"full_name" validate:"required"`
	Phone    *string `json:"phone,omitempty" validate:"omitempty,vn_phone"`
}

func TestFieldErrorsReportsJSONFieldNames(t *testing.T) {
	t.Parallel()

	err := validator.Struct(profileRequest{Phone: strPtr("0912")})

	require.Error(t, err)
	assert.Equal(t, map[string]string{
		"full_name": "required",
		"phone":     "vn_phone",
	}, validator.FieldErrors(err))
}

// The full name of an operational user is its identity in the sidebar and in
// the admin table: absent means "leave it alone", empty would mean "erase it".
func TestStruct_FullNameCannotBeEmptied(t *testing.T) {
	t.Parallel()

	t.Run("self profile keeps an omitted name", func(t *testing.T) {
		t.Parallel()
		if err := validator.Struct(dto.UpdateMeRequest{}); err != nil {
			t.Fatalf("expected an omitted full_name to pass, got: %v", err)
		}
	})

	t.Run("self profile refuses an empty name", func(t *testing.T) {
		t.Parallel()
		empty := ""
		err := validator.Struct(dto.UpdateMeRequest{FullName: &empty})
		if err == nil || !strings.Contains(err.Error(), "full_name") {
			t.Fatalf("expected a full_name error, got: %v", err)
		}
	})

	t.Run("role update requires a name", func(t *testing.T) {
		t.Parallel()
		err := validator.Struct(dto.UpdateUserRoleRequest{Role: "staff"})
		if err == nil || !strings.Contains(err.Error(), "full_name") {
			t.Fatalf("expected a full_name error, got: %v", err)
		}
	})

	t.Run("role update accepts a name", func(t *testing.T) {
		t.Parallel()
		if err := validator.Struct(dto.UpdateUserRoleRequest{
			Role:     "staff",
			FullName: "Mai Tran",
		}); err != nil {
			t.Fatalf("expected a named role update to pass, got: %v", err)
		}
	})
}
