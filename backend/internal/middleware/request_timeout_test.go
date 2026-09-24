package middleware

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestTimeout(t *testing.T) {
	t.Parallel()

	run := func(t *testing.T, timeout time.Duration, handler fiber.Handler) {
		t.Helper()
		app := fiber.New()
		app.Use(RequestTimeout(timeout))
		app.Get("/", handler)

		request := httptest.NewRequestWithContext(t.Context(), fiber.MethodGet, "/", nil)
		resp, err := app.Test(request, 2000)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
	}

	t.Run("cancels_the_work_a_handler_still_has_in_flight", func(t *testing.T) {
		t.Parallel()
		var err error
		run(t, 50*time.Millisecond, func(c *fiber.Ctx) error {
			<-c.UserContext().Done()
			err = c.UserContext().Err()
			return c.SendStatus(fiber.StatusOK)
		})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("leaves_a_handler_alone_when_no_deadline_is_configured", func(t *testing.T) {
		t.Parallel()
		var deadlineSet bool
		run(t, 0, func(c *fiber.Ctx) error {
			_, deadlineSet = c.UserContext().Deadline()
			return c.SendStatus(fiber.StatusOK)
		})
		assert.False(t, deadlineSet)
	})

	t.Run("a_fast_handler_keeps_its_deadline_unused", func(t *testing.T) {
		t.Parallel()
		var remaining time.Duration
		run(t, time.Second, func(c *fiber.Ctx) error {
			deadline, ok := c.UserContext().Deadline()
			require.True(t, ok)
			remaining = time.Until(deadline)
			return c.SendStatus(fiber.StatusOK)
		})
		assert.Positive(t, remaining)
	})
}
