package utils

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// phoneCases is the shared fixture both the Go and the TypeScript rule are held to.
type phoneCases struct {
	Normalize []struct {
		Input  string `json:"input"`
		Stored string `json:"stored"`
	} `json:"normalize"`
	Reject []string `json:"reject"`
}

func loadPhoneCases(t *testing.T) phoneCases {
	t.Helper()
	// A rooted file system: the read cannot leave contracts/.
	raw, err := fs.ReadFile(os.DirFS(filepath.Join("..", "..", "..", "..", "contracts")), "vietnam-phone-cases.json")
	require.NoError(t, err)
	var cases phoneCases
	require.NoError(t, json.Unmarshal(raw, &cases))
	require.NotEmpty(t, cases.Normalize)
	require.NotEmpty(t, cases.Reject)
	return cases
}

func TestNormalizeVietnamPhone(t *testing.T) {
	t.Parallel()
	cases := loadPhoneCases(t)

	t.Run("stores_every_accepted_spelling_in_one_form", func(t *testing.T) {
		t.Parallel()
		for _, c := range cases.Normalize {
			got, ok := NormalizeVietnamPhone(c.Input)
			assert.True(t, ok, c.Input)
			assert.Equal(t, c.Stored, got, c.Input)
		}
	})

	t.Run("rejects_what_is_not_a_vietnam_number", func(t *testing.T) {
		t.Parallel()
		for _, in := range cases.Reject {
			_, ok := NormalizeVietnamPhone(in)
			assert.False(t, ok, in)
		}
	})
}

func TestNormalizeVietnamPhonePtr(t *testing.T) {
	t.Parallel()

	t.Run("absent_stays_absent", func(t *testing.T) {
		t.Parallel()
		got, ok := NormalizeVietnamPhonePtr(nil)
		assert.True(t, ok)
		assert.Nil(t, got)
	})

	t.Run("blank_is_a_request_to_clear", func(t *testing.T) {
		t.Parallel()
		in := "   "
		got, ok := NormalizeVietnamPhonePtr(&in)
		assert.True(t, ok)
		require.NotNil(t, got)
		assert.Equal(t, "", *got)
	})

	t.Run("valid_number_is_normalized", func(t *testing.T) {
		t.Parallel()
		in := "0912 345 678"
		got, ok := NormalizeVietnamPhonePtr(&in)
		assert.True(t, ok)
		require.NotNil(t, got)
		assert.Equal(t, "+84912345678", *got)
	})

	t.Run("invalid_number_fails_instead_of_passing_through", func(t *testing.T) {
		t.Parallel()
		in := "!!!!!!"
		got, ok := NormalizeVietnamPhonePtr(&in)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}
