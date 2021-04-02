package doh

import (
	"strconv"
	"strings"
	"time"

	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/provider"
)

type ServerSettings struct {
	Resolver  ResolverSettings
	Port      uint16
	Cache     cache.Settings
	Blacklist blacklist.Settings
}

type ResolverSettings struct {
	DoHProviders []provider.Provider
	SelfDNS      SelfDNS
	Timeout      time.Duration
	IPv6         bool
}

type SelfDNS struct {
	// for the internal HTTP client to resolve the DoH url hostname.
	DoTProviders []provider.Provider
	DNSProviders []provider.Provider
}

func (s *ServerSettings) setDefaults() {
	s.Resolver.setDefaults()

	if s.Port == 0 {
		const defaultPort = 53
		s.Port = defaultPort
	}

	if string(s.Cache.Type) == "" {
		s.Cache.Type = cache.NOOP
	}
}

func (s *ResolverSettings) setDefaults() {
	s.SelfDNS.setDefaults()

	if len(s.DoHProviders) == 0 {
		s.DoHProviders = []provider.Provider{provider.Cloudflare()}
	}

	if s.Timeout == 0 {
		const defaultTimeout = 5 * time.Second
		s.Timeout = defaultTimeout
	}
}

func (s *SelfDNS) setDefaults() {
	if len(s.DoTProviders) == 0 {
		s.DoTProviders = []provider.Provider{provider.Cloudflare()}
	}
	// No default DNS fallback server for the internal HTTP client
	// to avoid leaking we are using a DoH server.
}

const (
	subSection = " |--"
	indent     = "    " // used if lines already contain the subSection
)

func (s *ServerSettings) String() string {
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *ResolverSettings) String() string {
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *SelfDNS) String() string {
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *ServerSettings) Lines(indent, subSection string) (lines []string) {
	lines = append(lines, subSection+"Resolver:")
	for _, line := range s.Resolver.Lines(indent, subSection) {
		lines = append(lines, indent+line)
	}

	lines = append(lines,
		subSection+"Listening port: "+strconv.Itoa(int(s.Port)))

	lines = append(lines, subSection+"Caching:")
	for _, line := range s.Cache.Lines(indent, subSection) {
		lines = append(lines, indent+line)
	}

	lines = append(lines, subSection+"Blacklist:")
	for _, line := range s.Blacklist.Lines(indent, subSection) {
		lines = append(lines, indent+line)
	}

	return lines
}

func (s *ResolverSettings) Lines(indent, subSection string) (lines []string) {
	lines = append(lines, subSection+"DNS over HTTPS providers:")
	for _, provider := range s.DoHProviders {
		lines = append(lines, indent+subSection+provider.String())
	}

	lines = append(lines, subSection+"Internal DNS:")
	for _, line := range s.SelfDNS.Lines(indent, subSection) {
		lines = append(lines, indent+line)
	}

	lines = append(lines,
		subSection+"Query timeout: "+s.Timeout.String())

	connectOver := "IPv4"
	if s.IPv6 {
		connectOver = "IPv6"
	}
	lines = append(lines, subSection+"Connecting over: "+connectOver)

	return lines
}

func (s *SelfDNS) Lines(indent, subSection string) (lines []string) {
	if len(s.DoTProviders) > 0 {
		lines = append(lines, subSection+"DNS over TLS providers:")
		for _, provider := range s.DoTProviders {
			lines = append(lines, indent+subSection+provider.String())
		}
	}

	if len(s.DNSProviders) > 0 {
		lines = append(lines, subSection+"Fallback plaintext DNS servers:")
		for _, provider := range s.DNSProviders {
			lines = append(lines, indent+subSection+provider.String())
		}
	}

	return lines
}
