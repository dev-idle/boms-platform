package config

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePrefixes(t *testing.T) {
	t.Parallel()

	t.Run("reads_addresses_and_ranges", func(t *testing.T) {
		t.Parallel()
		got, err := parsePrefixes("realtime.trusted_proxies", []string{"10.0.0.7", "192.168.1.9/24", "::ffff:10.0.0.8", "2001:db8::1"})
		require.NoError(t, err)
		assert.Equal(t, []netip.Prefix{
			netip.MustParsePrefix("10.0.0.7/32"),
			netip.MustParsePrefix("192.168.1.0/24"),
			netip.MustParsePrefix("10.0.0.8/32"),
			netip.MustParsePrefix("2001:db8::1/128"),
		}, got)
	})

	t.Run("names_the_setting_it_cannot_read", func(t *testing.T) {
		t.Parallel()
		_, err := parsePrefixes("realtime.trusted_proxies", []string{"edge.example.com"})
		require.ErrorContains(t, err, `realtime.trusted_proxies: "edge.example.com"`)
	})
}
