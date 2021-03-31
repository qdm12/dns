package doh

import (
	"time"

	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/provider"
)

type Settings struct {
	DoHServers []provider.DoHServer
	SelfDNS    SelfDNS
	Timeout    time.Duration
	Port       uint16
	IPv6       bool
	Cache      cache.Settings
	Blacklist  blacklist.Settings
}

type SelfDNS struct {
	// for the internal HTTP client to resolve the DoH url hostname.
	DoTServers []provider.DoTServer
	DNSServers []provider.DNSServer
}

func (s *Settings) setDefaults() {
	defaultProviders := []provider.Provider{provider.Cloudflare()}

	if len(s.DoHServers) == 0 {
		s.DoHServers = make([]provider.DoHServer, len(defaultProviders))
		for i := range defaultProviders {
			s.DoHServers[i] = defaultProviders[i].DoH()
		}
	}

	if len(s.SelfDNS.DoTServers) == 0 {
		s.SelfDNS.DoTServers = make([]provider.DoTServer, len(defaultProviders))
		for i := range defaultProviders {
			s.SelfDNS.DoTServers[i] = defaultProviders[i].DoT()
		}
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

// SetProviders set the DoH, DoT and fallback DNS servers settings for
// the providers given.
func (s *Settings) SetProviders(first provider.Provider, providers ...provider.Provider) {
	providers = append(providers, first)
	s.DoHServers = make([]provider.DoHServer, len(providers))
	s.SelfDNS.DoTServers = make([]provider.DoTServer, len(providers))
	s.SelfDNS.DNSServers = make([]provider.DNSServer, len(providers))
	for i := range providers {
		s.DoHServers[i] = providers[i].DoH()
		s.SelfDNS.DoTServers[i] = providers[i].DoT()
		s.SelfDNS.DNSServers[i] = providers[i].DNS()
	}
}
