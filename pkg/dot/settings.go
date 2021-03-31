package dot

import (
	"time"

	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/provider"
)

type Settings struct {
	DoTServers []provider.DoTServer
	DNSServers []provider.DNSServer
	Timeout    time.Duration
	Port       uint16
	IPv6       bool
	Cache      cache.Settings
	Blacklist  blacklist.Settings
}

func (s *Settings) setDefaults() {
	defaultProviders := []provider.Provider{provider.Cloudflare()}

	if len(s.DoTServers) == 0 {
		s.DoTServers = make([]provider.DoTServer, len(defaultProviders))
		for i := range defaultProviders {
			s.DoTServers[i] = defaultProviders[i].DoT()
		}
	}

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

// SetProviders set the DoT servers settings for the providers given.
func (s *Settings) SetProviders(first provider.Provider, providers ...provider.Provider) {
	providers = append(providers, first)
	s.DoTServers = make([]provider.DoTServer, len(providers))
	for i := range providers {
		s.DoTServers[i] = providers[i].DoT()
	}
}

// SetDNSFallback set the plaintext DNS fallback servers settings for the providers given.
func (s *Settings) SetDNSFallback(first provider.Provider, providers ...provider.Provider) {
	providers = append(providers, first)
	s.DNSServers = make([]provider.DNSServer, len(providers))
	for i := range providers {
		s.DNSServers[i] = providers[i].DNS()
	}
}
