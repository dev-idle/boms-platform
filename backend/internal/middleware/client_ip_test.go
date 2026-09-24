package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientIP(t *testing.T) {
	t.Parallel()

	run := func(t *testing.T, header string) string {
		t.Helper()
		app := fiber.New()
		var seen string
		app.Get("/", func(c *fiber.Ctx) error {
			seen = ClientIP(c)
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequestWithContext(t.Context(), fiber.MethodGet, "/", nil)
		if header != "" {
			req.Header.Set(ClientIPHeader, header)
		}
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		return seen
	}

	t.Run("reads_the_address_the_bff_stamped", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "203.0.113.7", run(t, "203.0.113.7"))
		assert.Equal(t, "2001:db8::1", run(t, " 2001:db8::1 "))
	})

	t.Run("falls_back_to_the_socket_when_the_header_cannot_be_trusted", func(t *testing.T) {
		t.Parallel()
		// A caller that reaches the API directly must not be able to pick its own
		// rate-limit bucket, or write a chosen address into an audit row.
		for _, header := range []string{"", "not-an-ip", "203.0.113.7, 10.0.0.1", "'; DROP TABLE"} {
			assert.Equal(t, "0.0.0.0", run(t, header), header)
		}
	})
}
