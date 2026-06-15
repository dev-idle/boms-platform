package v1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/service/cloudinary"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerMediaHandler_GetCloudinaryUploadSignature(t *testing.T) {
	t.Parallel()

	t.Run("service_unavailable when cloudinary disabled", func(t *testing.T) {
		t.Parallel()
		handler := NewManagerMediaHandler(usecase.NewManagerMediaUsecase(config.CloudinaryConfig{}))
		app := fiber.New()
		app.Get("/", handler.GetCloudinaryUploadSignature)

		resp := requestManagerMediaHandler(t, app)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	})

	t.Run("ok when cloudinary configured", func(t *testing.T) {
		t.Parallel()
		handler := NewManagerMediaHandler(usecase.NewManagerMediaUsecase(config.CloudinaryConfig{
			CloudName:    "demo",
			APIKey:       "key",
			APISecret:    "secret",
			UploadFolder: "boms/products",
		}))
		app := fiber.New()
		app.Get("/", handler.GetCloudinaryUploadSignature)

		resp := requestManagerMediaHandler(t, app)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var envelope struct {
			Data struct {
				CloudName string `json:"cloud_name"`
				MaxBytes  int64  `json:"max_bytes"`
				Folder    string `json:"folder"`
				Signature string `json:"signature"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(body, &envelope))
		assert.Equal(t, "demo", envelope.Data.CloudName)
		assert.Equal(t, "boms/products", envelope.Data.Folder)
		assert.Equal(t, cloudinary.MaxProductImageBytes, envelope.Data.MaxBytes)
		assert.Len(t, envelope.Data.Signature, 40)
	})
}

func requestManagerMediaHandler(t *testing.T, app *fiber.App) *http.Response {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}
