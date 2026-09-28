package store

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

// The clock cases in contracts/pickup-rules-cases.json, which the frontend's
// clockTimeSchema reads too.
func TestClock_Contract(t *testing.T) {
	t.Parallel()
	// A rooted file system: the read cannot leave contracts/.
	raw, err := fs.ReadFile(os.DirFS(filepath.Join("..", "..", "..", "..", "contracts")), "pickup-rules-cases.json")
	require.NoError(t, err)
	var cases struct {
		Clock struct {
			Valid []struct {
				Text    string `json:"text"`
				Minutes int    `json:"minutes"`
			} `json:"valid"`
			Invalid []string `json:"invalid"`
		} `json:"clock"`
	}
	require.NoError(t, json.Unmarshal(raw, &cases))
	require.NotEmpty(t, cases.Clock.Valid)

	for _, tc := range cases.Clock.Valid {
		got, err := ParseClock(tc.Text)
		require.NoError(t, err, tc.Text)
		assert.Equal(t, time.Duration(tc.Minutes)*time.Minute, got, tc.Text)
		assert.Equal(t, tc.Text, FormatClock(got))
	}
	for _, text := range cases.Clock.Invalid {
		_, err := ParseClock(text)
		assert.Error(t, err, "%q", text)
	}
}
