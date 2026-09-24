package middleware

import (
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ClientIPHeader carries the browser's address from the BFF.
//
// Every `/api/v1` route sits behind RequireInternalSecret, so the BFF is the
// only caller the API ever has: `c.IP()` is therefore the BFF's own address —
// one bucket for every visitor in the rate limiter, and one address in every
// audit row. The BFF strips whatever the browser sent for this header and the
// forwarding headers, then stamps this one from its own trusted source.
const ClientIPHeader = "X-Client-IP"

// ClientIP is the address of the person making the request, for rate-limit keys
// and audit rows. It falls back to the socket address when the header is absent
// or not an address — a direct caller must not be able to widen its own bucket
// by sending a malformed value.
func ClientIP(c *fiber.Ctx) string {
	candidate := strings.TrimSpace(c.Get(ClientIPHeader))
	if candidate == "" {
		return c.IP()
	}
	if net.ParseIP(candidate) == nil {
		return c.IP()
	}
	return candidate
}
