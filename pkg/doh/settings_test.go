package doh

import (
	"testing"
	"time"

	cache "github.com/qdm12/dns/pkg/cache/noop"
	metrics "github.com/qdm12/dns/pkg/doh/metrics/noop"
	"github.com/qdm12/dns/pkg/filter"
	log "github.com/qdm12/dns/pkg/log/noop"
	"github.com/qdm12/dns/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_ServerSettings_setDefaults(t *testing.T) {
	t.Parallel()

	cache := cache.New(cache.Settings{})
	filter := filter.NewMap(filter.Settings{})
	metrics := metrics.New()
	logger := log.New()

	s := ServerSettings{
		Cache:   cache,
		Filter:  filter,
		Metrics: metrics,
		Logger:  logger,
		Resolver: ResolverSettings{
			Warner:  logger,
			Metrics: metrics,
		},
	}
	s.setDefaults()

	// Check this otherwise things will blow up if no option is passed.
	assert.GreaterOrEqual(t, len(s.Resolver.DoHProviders), 1)
	assert.GreaterOrEqual(t, len(s.Resolver.SelfDNS.DoTProviders), 1)
	assert.Empty(t, s.Resolver.SelfDNS.DNSProviders)
	assert.GreaterOrEqual(t, int64(s.Resolver.Timeout), int64(time.Millisecond))

	expectedSettings := ServerSettings{
		Cache:   cache,
		Filter:  filter,
		Metrics: metrics,
		Logger:  logger,
		Resolver: ResolverSettings{
			DoHProviders: []provider.Provider{provider.Cloudflare()},
			SelfDNS: SelfDNS{
				DoTProviders: []provider.Provider{provider.Cloudflare()},
				Timeout:      5 * time.Second,
				IPv6:         false,
			},
			Timeout: 5 * time.Second,
			Warner:  logger,
			Metrics: metrics,
		},
		Port: 53,
	}
	assert.Equal(t, expectedSettings, s)
}

func Test_ServerSettings_Lines(t *testing.T) {
	t.Parallel()

	s := ServerSettings{}
	s.setDefaults()

	lines := s.Lines(indent, subSection)

	expectedLines := []string{
		" |--Listening port: 53",
		" |--Resolver:",
		"     |--Query timeout: 5s",
		"     |--DNS over HTTPS providers:",
		"         |--Cloudflare",
		"     |--Internal DNS:",
		"         |--Connecting using IPv4 DNS addresses",
		"         |--Query timeout: 5s",
		"         |--DNS over TLS providers:",
		"             |--Cloudflare",
	}
	assert.Equal(t, expectedLines, lines)
}
