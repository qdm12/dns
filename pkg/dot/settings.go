package dot

import (
	"time"

	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/provider"
)

type Settings struct {
	DoTProviders []provider.Provider
	DNSProviders []provider.Provider
	Timeout      time.Duration
	Port         uint16
	IPv6         bool
	Cache        cache.Settings
	Blacklist    blacklist.Settings
}

func (s *Settings) setDefaults() {
	if len(s.DoTProviders) == 0 {
		s.DoTProviders = []provider.Provider{provider.Cloudflare()}
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
