package dot

import (
	"strconv"
	"strings"
	"time"

	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/middlewares/log"
	"github.com/qdm12/dns/pkg/provider"
)

type ServerSettings struct {
	Resolver  ResolverSettings
	Port      uint16
	Log       log.Settings
	Cache     cache.Settings
	Blacklist blacklist.Settings
}

type ResolverSettings struct {
	DoTProviders []provider.Provider
	DNSProviders []provider.Provider
	Timeout      time.Duration
	IPv6         bool
}

func (s *ServerSettings) setDefaults() {
	s.Resolver.setDefaults()

	if s.Port == 0 {
		const defaultPort = 53
		s.Port = defaultPort
	}

	// Cache defaults to disabled, see pkg/cache/settings.go
	s.Cache.SetDefaults()
}

func (s *ResolverSettings) setDefaults() {
	if len(s.DoTProviders) == 0 {
		s.DoTProviders = []provider.Provider{provider.Cloudflare()}
	}

	if s.Timeout == 0 {
		const defaultTimeout = 5 * time.Second
		s.Timeout = defaultTimeout
	}
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

	connectOver := "IPv4"
	if s.IPv6 {
		connectOver = "IPv6"
	}
	lines = append(lines, subSection+"Connecting over: "+connectOver)

	return lines
}
