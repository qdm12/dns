package doh

import (
	"strconv"
	"strings"
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

func (s *Settings) String() string {
	const (
		subSection = " |--"
		indent     = "    " // used if lines already contain the subSection
	)
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *Settings) Lines(indent, subSection string) (lines []string) {
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

	lines = append(lines,
		subSection+"Listening port: "+strconv.Itoa(int(s.Port)))

	connectOver := "IPv4"
	if s.IPv6 {
		connectOver = "IPv6"
	}
	lines = append(lines, subSection+"Connecting over: "+connectOver)

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
