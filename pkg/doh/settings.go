package doh

import (
	"time"

	"github.com/qdm12/dns/pkg/provider"
)

type settings struct {
	providers  []provider.Provider // for the internal HTTP client
	dohServers []provider.DoHServer
	timeout    time.Duration
	ipv6       bool
}

func defaultSettings() (settings settings) {
	settings.providers = []provider.Provider{provider.Cloudflare()}
	settings.dohServers = make([]provider.DoHServer, len(settings.providers))
	for i := range settings.providers {
		settings.dohServers[i] = settings.providers[i].DoH()
	}

	const defaultTimeout = 5 * time.Second
	settings.timeout = defaultTimeout

	return settings
}
