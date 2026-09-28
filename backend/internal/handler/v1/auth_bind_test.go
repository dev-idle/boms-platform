package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/middleware"
	"github.com/boms/backend/internal/shared/response"
)

// A body the binder cannot read is the client's mistake: it must come back as
// 400 validation_error before the usecase is reached (it is nil here), whatever
// the framework's binder returns for that case.
func TestRegister_RejectsUnreadableBodies(t *testing.T) {
	t.Parallel()

	handler := NewAuthHandler(nil, &config.Config{})
	cases := []struct {
		name        string
		contentType string
		body        string
	}{
		{name: "malformed_json", contentType: fiber.MIMEApplicationJSON, body: `{"email":`},
		{name: "missing_content_type", body: `{"email":"mai@example.com","password":"x"}`},
		{name: "empty_body", contentType: fiber.MIMEApplicationJSON},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zap.NewNop())})
			app.Post("/", handler.Register)

			req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set(fiber.HeaderContentType, tc.contentType)
			}
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			var env response.Envelope
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
			require.NotNil(t, env.Error)
			assert.Equal(t, "validation_error", env.Error.Code)
			// The binder, not the validator, turned it away: the body never reached a struct.
			assert.Equal(t, "invalid_body", env.Error.Details["reason"])
		})
	}
}
