package errors_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The API contract defines every client-visible error code twice: here in Go,
// and in the frontend's ApiErrorCode, which is what the UI branches on. A code
// the backend can send but the frontend does not know falls through to a generic
// toast. This test fails the backend build the moment the two drift apart.
//
// The frontend may know extra codes — timeouts and malformed responses are raised
// in the browser, never by the API — so only one direction is checked.

var (
	backendCodePattern  = regexp.MustCompile(`New\(http\.Status\w+,\s*"([a-z_]+)"`)
	frontendBlock       = regexp.MustCompile(`(?s)export const ApiErrorCode = \{(.*?)\} as const`)
	frontendCodePattern = regexp.MustCompile(`:\s*"([a-z_]+)"`)
)

func TestEveryBackendErrorCodeIsMirroredInTheFrontend(t *testing.T) {
	t.Parallel()

	// Rooted file systems: the reads below cannot leave the repository.
	backend := backendErrorCodes(t, os.DirFS(filepath.Join("..", "..")))
	frontend := frontendErrorCodes(t, os.DirFS(filepath.Join("..", "..", "..", "..", "frontend")), "src/lib/errors/api-error.ts")

	var missing []string
	for code := range backend {
		if !frontend[code] {
			missing = append(missing, code)
		}
	}
	sort.Strings(missing)
	require.Empty(t, missing, "add these codes to ApiErrorCode in frontend/src/lib/errors/api-error.ts")
}

// backendErrorCodes collects the code of every AppError constructed in
// non-test Go source under internal/.
func backendErrorCodes(t *testing.T, internal fs.FS) map[string]bool {
	t.Helper()
	codes := map[string]bool{}
	err := fs.WalkDir(internal, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, readErr := fs.ReadFile(internal, path)
		if readErr != nil {
			return readErr
		}
		for _, m := range backendCodePattern.FindAllSubmatch(src, -1) {
			codes[string(m[1])] = true
		}
		return nil
	})
	require.NoError(t, err)
	// A pattern that silently matches nothing would pass vacuously.
	require.Greater(t, len(codes), 20, "backend error-code scan found too few codes")
	return codes
}

func frontendErrorCodes(t *testing.T, frontend fs.FS, path string) map[string]bool {
	t.Helper()
	src, err := fs.ReadFile(frontend, path)
	require.NoError(t, err)
	block := frontendBlock.FindSubmatch(src)
	require.NotNil(t, block, "ApiErrorCode block not found in %s", path)
	codes := map[string]bool{}
	for _, m := range frontendCodePattern.FindAllSubmatch(block[1], -1) {
		codes[string(m[1])] = true
	}
	require.NotEmpty(t, codes)
	return codes
}
