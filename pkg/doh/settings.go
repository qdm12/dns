package doh

import (
	"time"

	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/provider"
)

type Settings struct {
	DoHProviders []provider.Provider
	SelfDNS      SelfDNS
	Timeout      time.Duration
	Port         uint16
	IPv6         bool
	Cache        cache.Settings
	Blacklist    blacklist.Settings
}

type SelfDNS struct {
	// for the internal HTTP client to resolve the DoH url hostname.
	DoTProviders []provider.Provider
	DNSProviders []provider.Provider
}

func (s *Settings) setDefaults() {
	if len(s.DoHProviders) == 0 {
		s.DoHProviders = []provider.Provider{provider.Cloudflare()}
	}

	if len(s.SelfDNS.DoTProviders) == 0 {
		s.SelfDNS.DoTProviders = []provider.Provider{provider.Cloudflare()}
	}

	// No default DNS fallback server for the internal HTTP client
	// to avoid leaking we are using a DoH server.

	if s.Port == 0 {
		const defaultPort = 53
		s.Port = defaultPort
	}

	if s.Timeout == 0 {
		const defaultTimeout = 5 * time.Second
		s.Timeout = defaultTimeout
	}

	if string(s.Cache.Type) == "" {
		s.Cache.Type = cache.NOOP
	}
}
