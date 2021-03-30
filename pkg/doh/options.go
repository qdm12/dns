package doh

import (
	"net"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/provider"
)

type Option func(s *settings)

func Port(port uint16) Option {
	return func(s *settings) {
		s.port = port
	}
}

func Providers(first provider.Provider, providers ...provider.Provider) Option {
	providers = append(providers, first)
	dohServers := make([]provider.DoHServer, len(providers))
	for i := range providers {
		dohServers[i] = providers[i].DoH()
	}

	return func(s *settings) {
		s.providers = providers
		s.dohServers = dohServers
	}
}

func Timeout(timeout time.Duration) Option {
	return func(s *settings) {
		s.timeout = timeout
	}
}

func IPv4() Option {
	return func(s *settings) { s.ipv6 = false }
}

func IPv6() Option {
	return func(s *settings) { s.ipv6 = true }
}

func WithCache(cacheType cache.Type, options ...cache.Option) Option {
	return func(s *settings) {
		s.cacheType = cacheType
		s.cacheOptions = options
	}
}

func BlockHostnames(hostnames []string) Option {
	fqdnHostnames := make([]string, len(hostnames))
	for i := range hostnames {
		fqdnHostnames[i] = dns.Fqdn(hostnames[i])
	}
	return func(s *settings) {
		s.blacklist.fqdnHostnames = fqdnHostnames
	}
}

func BlockIPs(ips []net.IP) Option {
	return func(s *settings) {
		s.blacklist.ips = ips
	}
}
