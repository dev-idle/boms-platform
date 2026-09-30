package policy

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

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

func TestRequireAccepted(t *testing.T) {
	t.Parallel()

	require.NoError(t, RequireAccepted(TermsVersion))
	for _, version := range []string{"", "2026-01-01", " " + TermsVersion} {
		require.ErrorIs(t, RequireAccepted(version), ErrTermsNotAccepted, "version %q", version)
	}
}
