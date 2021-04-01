package dot

import (
	"strconv"
	"strings"
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

func (s *Settings) String() string {
	const (
		subSection = " |--"
		indent     = "    " // used if lines already contain the subSection
	)
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *Settings) Lines(indent, subSection string) (lines []string) {
	lines = append(lines, subSection+"DNS over TLS providers:")
	for _, provider := range s.DoTProviders {
		lines = append(lines, indent+subSection+provider.String())
	}

	lines = append(lines, subSection+"Fallback plaintext DNS providers:")
	for _, provider := range s.DNSProviders {
		lines = append(lines, indent+subSection+provider.String())
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
