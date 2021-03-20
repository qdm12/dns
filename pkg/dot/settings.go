package dot

import (
	"time"

	"github.com/qdm12/dns/pkg/provider"
)

type settings struct {
	dotServers []provider.DoTServer
	dnsServers []provider.DNSServer
	timeout    time.Duration
	ipv6       bool
}

func defaultSettings() (settings settings) {
	providers := []provider.Provider{provider.Cloudflare()}
	settings.dotServers = make([]provider.DoTServer, len(providers))
	for i := range providers {
		settings.dotServers[i] = providers[i].DoT()
	}

	const defaultTimeout = 5 * time.Second
	settings.timeout = defaultTimeout

	return settings
}
