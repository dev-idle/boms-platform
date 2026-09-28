package realtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAdmission(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 28, 9, 0, 0, 0, time.UTC)
	first := netip.MustParseAddr("203.0.113.7")
	second := netip.MustParseAddr("203.0.113.8")

	t.Run("one_address_gets_only_its_share", func(t *testing.T) {
		t.Parallel()
		a := newAdmission(100, 1)

		assert.True(t, a.allow(first, now))
		assert.True(t, a.allow(first, now), "a burst of twice the rate")
		assert.False(t, a.allow(first, now))
		assert.True(t, a.allow(second, now), "another address keeps its own share")
		assert.True(t, a.allow(first, now.Add(time.Second)), "the share refills")
	})

	t.Run("the_process_caps_every_address_together", func(t *testing.T) {
		t.Parallel()
		a := newAdmission(1, 1)

		assert.True(t, a.allow(first, now))
		assert.True(t, a.allow(second, now))
		assert.False(t, a.allow(netip.MustParseAddr("203.0.113.9"), now))
	})

	t.Run("forgets_addresses_that_went_quiet", func(t *testing.T) {
		t.Parallel()
		a := newAdmission(100, 10)
		a.allow(first, now)
		a.allow(second, now.Add(50*time.Second))

		a.allow(second, now.Add(addressIdle+time.Second))

		assert.NotContains(t, a.perAddr, first)
		assert.Contains(t, a.perAddr, second)
	})
}

func TestClientAddress(t *testing.T) {
	t.Parallel()
	edge := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}

	request := func(peer string, forwarded ...string) *http.Request {
		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ws", nil)
		r.RemoteAddr = peer
		for _, value := range forwarded {
			r.Header.Add("X-Forwarded-For", value)
		}
		return r
	}

	cases := map[string]struct {
		r    *http.Request
		want string
	}{
		"an_untrusted_peer_is_the_client":      {request("198.51.100.4:5000", "203.0.113.7"), "198.51.100.4"},
		"a_proxy_names_the_client":             {request("10.0.0.2:5000", "203.0.113.7"), "203.0.113.7"},
		"a_forged_hop_is_left_of_the_real_one": {request("10.0.0.2:5000", "1.2.3.4, 203.0.113.7"), "203.0.113.7"},
		"proxies_in_the_chain_are_skipped":     {request("10.0.0.2:5000", "203.0.113.7", "10.0.0.9"), "203.0.113.7"},
		"a_malformed_hop_keeps_the_proxy":      {request("10.0.0.2:5000", "not-an-ip"), "10.0.0.2"},
		"a_proxy_without_the_header":           {request("10.0.0.2:5000"), "10.0.0.2"},
		"a_mapped_ipv4_peer":                   {request("[::ffff:198.51.100.4]:5000"), "198.51.100.4"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, netip.MustParseAddr(tc.want), clientAddress(tc.r, edge))
		})
	}
}
