package doh

import (
	"testing"
	"time"

	metrics "github.com/qdm12/dns/v2/pkg/doh/metrics/noop"
	log "github.com/qdm12/dns/v2/pkg/log/noop"
	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_ServerSettings_SetDefaults(t *testing.T) {
	t.Parallel()

	metrics := metrics.New()
	logger := log.New()

	s := ServerSettings{
		Middlewares: []Middleware{},
		Logger:      logger,
		Resolver: ResolverSettings{
			Metrics: metrics,
		},
	}
	s.SetDefaults()

	// Check this otherwise things will blow up if no option is passed.
	assert.GreaterOrEqual(t, len(s.Resolver.UpstreamResolvers), 1)
	assert.Equal(t, "ipv4", s.Resolver.IPVersion)
	assert.GreaterOrEqual(t, int64(s.Resolver.Timeout), int64(time.Millisecond))

	expectedSettings := ServerSettings{
		Middlewares: []Middleware{},
		Logger:      logger,
		Resolver: ResolverSettings{
			UpstreamResolvers: []provider.Provider{provider.Cloudflare()},
			IPVersion:         "ipv4",
			Timeout:           5 * time.Second,
			Metrics:           metrics,
		},
		ListeningAddress: ptrTo(":53"),
	}
	assert.Equal(t, expectedSettings, s)
}

func Test_ServerSettings_String(t *testing.T) {
	t.Parallel()

	settings := ServerSettings{}
	settings.SetDefaults()

	s := settings.String()

	const expected = `DoH server settings:
├── Listening address: :53
└── DoH resolver settings:
    ├── Upstream resolvers:
    |   └── Cloudflare
    ├── Connecting over ipv4
    └── Query timeout: 5s`
	assert.Equal(t, expected, s)
}
