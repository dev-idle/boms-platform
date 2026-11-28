package middleware

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLogger(t *testing.T) {
	t.Parallel()

	run := func(t *testing.T, path string) (int, observer.LoggedEntry) {
		t.Helper()
		core, logs := observer.New(zapcore.InfoLevel)
		app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zap.NewNop())})
		app.Use(RequestLogger(zap.New(core)))
		app.Get("/created", func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusCreated) })
		app.Get("/broken", func(fiber.Ctx) error { return errors.New("boom") })

		resp, err := app.Test(httptest.NewRequestWithContext(t.Context(), fiber.MethodGet, path, nil))
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		entries := logs.FilterMessage("http_request").All()
		require.Len(t, entries, 1)
		return resp.StatusCode, entries[0]
	}

	t.Run("a_handled_request_is_logged_at_info_with_its_status", func(t *testing.T) {
		t.Parallel()
		status, entry := run(t, "/created")
		assert.Equal(t, fiber.StatusCreated, status)
		assert.Equal(t, zapcore.InfoLevel, entry.Level)
		assert.EqualValues(t, fiber.StatusCreated, entry.ContextMap()["status"])
	})

	t.Run("an_unknown_route_is_logged_with_the_404_the_client_gets", func(t *testing.T) {
		t.Parallel()
		status, entry := run(t, "/missing")
		assert.Equal(t, fiber.StatusNotFound, status)
		assert.Equal(t, zapcore.WarnLevel, entry.Level)
		assert.EqualValues(t, fiber.StatusNotFound, entry.ContextMap()["status"])
	})

	t.Run("an_unmapped_error_is_logged_as_the_500_the_client_gets", func(t *testing.T) {
		t.Parallel()
		status, entry := run(t, "/broken")
		assert.Equal(t, fiber.StatusInternalServerError, status)
		assert.Equal(t, zapcore.ErrorLevel, entry.Level)
		assert.EqualValues(t, fiber.StatusInternalServerError, entry.ContextMap()["status"])
	})
}
