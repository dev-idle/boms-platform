package policy

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The version is contracts/terms-version.json, which the frontend's constant
// is tested against too: the two sides must name the same policies.
func TestTermsVersion_MatchesTheContract(t *testing.T) {
	t.Parallel()
	raw, err := fs.ReadFile(os.DirFS(filepath.Join("..", "..", "..", "..", "contracts")), "terms-version.json")
	require.NoError(t, err)
	var contract struct {
		Version string `json:"version"`
	}
	require.NoError(t, json.Unmarshal(raw, &contract))

	assert.Equal(t, contract.Version, TermsVersion)
}

func TestTerms_Check(t *testing.T) {
	t.Parallel()
	servedSince := time.Date(2026, 11, 25, 8, 0, 0, 0, time.UTC)
	terms := NewTerms(servedSince)

	require.NoError(t, terms.Check(TermsVersion, servedSince.Add(30*24*time.Hour)), "the current version, whenever")
	require.NoError(t, terms.Check(PreviousTermsVersion, servedSince.Add(PreviousTermsOverlap-time.Second)),
		"a page opened before the change")
	require.ErrorIs(t, terms.Check(PreviousTermsVersion, servedSince.Add(PreviousTermsOverlap)), ErrTermsNotAccepted,
		"a page left open past the overlap")
	for _, version := range []string{"", "2026-01-01", " " + TermsVersion} {
		require.ErrorIs(t, terms.Check(version, servedSince), ErrTermsNotAccepted, "version %q", version)
	}
	assert.NotEqual(t, TermsVersion, PreviousTermsVersion, "the previous version moves on with each change")
}
