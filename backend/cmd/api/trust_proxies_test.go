package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Behind a trusted proxy the client address comes from X-Forwarded-For, but
// never as whatever text the caller put there: rate-limit buckets and audit
// rows are keyed by it. A real listener is used because c.IP() reads the
// socket's remote address to decide whether the proxy is trusted.
func TestTrustProxies_ReadsOnlyAValidatedForwardedAddress(t *testing.T) {
	t.Parallel()

	fcfg := fiber.Config{}
	trustProxies(&fcfg, []string{"127.0.0.1"})
	app := fiber.New(fcfg)
	app.Get("/ip", func(c fiber.Ctx) error { return c.SendString(c.IP()) })

	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true}) }()
	t.Cleanup(func() { _ = app.Shutdown() })

	ipFor := func(t *testing.T, forwardedFor string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+ln.Addr().String()+"/ip", nil)
		require.NoError(t, err)
		if forwardedFor != "" {
			req.Header.Set(fiber.HeaderXForwardedFor, forwardedFor)
		}
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return string(body)
	}

	cases := []struct {
		name         string
		forwardedFor string
		want         string
	}{
		{name: "takes_the_address_the_trusted_proxy_appended", forwardedFor: "6.6.6.6, 7.7.7.7", want: "7.7.7.7"},
		{name: "falls_back_to_the_socket_for_garbage", forwardedFor: "not-an-ip", want: "127.0.0.1"},
		{name: "falls_back_to_the_socket_without_the_header", want: "127.0.0.1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ipFor(t, tc.forwardedFor))
		})
	}
}

func TestTrustProxies_LeavesTheConfigAloneWithoutProxies(t *testing.T) {
	t.Parallel()
	fcfg := fiber.Config{}
	trustProxies(&fcfg, nil)
	assert.False(t, fcfg.TrustProxy)
	assert.False(t, fcfg.EnableIPValidation)
	assert.Empty(t, fcfg.ProxyHeader)
}
