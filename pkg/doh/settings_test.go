package doh

import (
	"net/url"
	"testing"
	"time"

	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_defaultSettings(t *testing.T) {
	t.Parallel()

	s := defaultSettings()

	// Check this otherwise things will blow up if no option is passed.
	assert.GreaterOrEqual(t, len(s.dohServers), 1)
	assert.GreaterOrEqual(t, len(s.providers), 1)
	assert.GreaterOrEqual(t, int64(s.timeout), int64(time.Millisecond))

	expectedSettings := settings{
		providers: []provider.Provider{provider.Cloudflare()},
		dohServers: []provider.DoHServer{{
			URL: &url.URL{
				Scheme: "https",
				Host:   "cloudflare-dns.com",
				Path:   "/dns-query",
			},
		}},
		timeout:   5 * time.Second,
		ipv6:      false,
		cacheType: cache.NOOP,
	}
	assert.Equal(t, expectedSettings, s)
}
