package doh

import (
	"testing"
	"time"

	metrics "github.com/qdm12/dns/v2/pkg/doh/metrics/noop"
	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_Settings_SetDefaults(t *testing.T) {
	t.Parallel()

	metrics := metrics.New()

	s := Settings{
		Metrics: metrics,
	}
	s.SetDefaults()

	// Check this otherwise things will blow up if no option is passed.
	assert.GreaterOrEqual(t, len(s.UpstreamResolvers), 1)
	assert.Equal(t, "ipv4", s.IPVersion)
	assert.GreaterOrEqual(t, int64(s.Timeout), int64(time.Millisecond))

	expectedSettings := Settings{
		UpstreamResolvers: []provider.Provider{provider.Cloudflare()},
		IPVersion:         "ipv4",
		Timeout:           5 * time.Second,
		Metrics:           metrics,
	}
	assert.Equal(t, expectedSettings, s)
}

func Test_Settings_String(t *testing.T) {
	t.Parallel()

	settings := Settings{}
	settings.SetDefaults()

	s := settings.String()

	const expected = `DoH resolver settings:
├── Upstream resolvers:
|   └── Cloudflare
├── Query timeout: 5s
└── Connecting over: ipv4`
	assert.Equal(t, expected, s)
}
