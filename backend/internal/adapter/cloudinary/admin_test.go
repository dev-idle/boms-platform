package cloudinary

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeCloudinary(t *testing.T, api http.HandlerFunc) *Admin {
	t.Helper()
	server := httptest.NewServer(api)
	t.Cleanup(server.Close)
	return &Admin{baseURL: server.URL, cloudName: "demo", apiKey: "key", apiSecret: "secret", http: server.Client()}
}

func TestAdmin_DeleteFolder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("deletes_everything_under_the_folder_and_its_cdn_copies", func(t *testing.T) {
		t.Parallel()
		var calls atomic.Int32
		admin := fakeCloudinary(t, func(w http.ResponseWriter, r *http.Request) {
			n := calls.Add(1)
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "/v1_1/demo/resources/image/upload", r.URL.Path)
			assert.Equal(t, "boms/references/u-1/", r.URL.Query().Get("prefix"), "the folder's own images, never a sibling's")
			assert.Equal(t, "true", r.URL.Query().Get("invalidate"))
			user, pass, ok := r.BasicAuth()
			assert.True(t, ok && user == "key" && pass == "secret")
			if n == 1 {
				assert.Empty(t, r.URL.Query().Get("next_cursor"))
				_, _ = io.WriteString(w, `{"deleted":{"boms/references/u-1/a":"deleted"},"partial":true,"next_cursor":"c2"}`)
				return
			}
			assert.Equal(t, "c2", r.URL.Query().Get("next_cursor"), "the next call goes on from the first")
			_, _ = io.WriteString(w, `{"deleted":{"boms/references/u-1/b":"deleted"},"partial":false}`)
		})

		require.NoError(t, admin.DeleteFolder(ctx, "boms/references/u-1"))
		assert.Equal(t, int32(2), calls.Load(), "asked again while images were left")
	})

	t.Run("an_empty_folder_is_done", func(t *testing.T) {
		t.Parallel()
		admin := fakeCloudinary(t, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"deleted":{},"partial":false}`)
		})

		require.NoError(t, admin.DeleteFolder(ctx, "boms/references/u-2"))
	})

	t.Run("a_refusal_is_an_error_naming_cloudinarys_reason", func(t *testing.T) {
		t.Parallel()
		admin := fakeCloudinary(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"error":{"message":"Invalid Signature"}}`)
		})

		err := admin.DeleteFolder(ctx, "boms/references/u-3")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "401")
		assert.Contains(t, err.Error(), "Invalid Signature")
		assert.NotContains(t, err.Error(), "secret", "the credentials stay out of errors")
	})
}
