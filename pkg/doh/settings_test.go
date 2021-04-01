package doh

import (
	"testing"
	"time"

	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_Settings_setDefaults(t *testing.T) {
	t.Parallel()

	s := Settings{}
	s.setDefaults()

	// Check this otherwise things will blow up if no option is passed.
	assert.GreaterOrEqual(t, len(s.DoHProviders), 1)
	assert.GreaterOrEqual(t, len(s.SelfDNS.DoTProviders), 1)
	assert.Empty(t, s.SelfDNS.DNSProviders)
	assert.GreaterOrEqual(t, int64(s.Timeout), int64(time.Millisecond))

	expectedSettings := Settings{
		DoHProviders: []provider.Provider{provider.Cloudflare()},
		SelfDNS: SelfDNS{
			DoTProviders: []provider.Provider{provider.Cloudflare()},
		},
		Timeout: 5 * time.Second,
		Port:    53,
		IPv6:    false,
		Cache: cache.Settings{
			Type: cache.NOOP,
		},
	}
	assert.Equal(t, expectedSettings, s)
}
