package dot

import (
	"fmt"
	"strings"
	"time"

	"github.com/qdm12/dns/pkg/cache"
	cachenoop "github.com/qdm12/dns/pkg/cache/noop"
	"github.com/qdm12/dns/pkg/dot/metrics"
	metricsnoop "github.com/qdm12/dns/pkg/dot/metrics/noop"
	"github.com/qdm12/dns/pkg/filter"
	filternoop "github.com/qdm12/dns/pkg/filter/noop"
	"github.com/qdm12/dns/pkg/log"
	lognoop "github.com/qdm12/dns/pkg/log/noop"
	logmiddleware "github.com/qdm12/dns/pkg/middlewares/log"
	"github.com/qdm12/dns/pkg/provider"
)

type ServerSettings struct {
	Resolver      ResolverSettings
	Port          uint16
	LogMiddleware logmiddleware.Settings
	// Cache is the cache to use in the server.
	// It defaults to a No-Op cache implementation with
	// a No-Op cache metrics implementation.
	Cache cache.Interface
	// Filter is the filter for DNS requests and responses.
	// It defaults to a No-Op filter implementation.
	Filter filter.Filter
	// Logger is the logger to log information.
	// It defaults to a No-Op logger implementation.
	Logger log.Logger
	// Metrics is the metrics interface to record metric data.
	// It defaults to a No-Op metrics implementation.
	Metrics metrics.Interface
}

type ResolverSettings struct {
	DoTProviders []provider.Provider
	DNSProviders []provider.Provider
	Timeout      time.Duration
	IPv6         bool
	// Warner is the warning logger to log dial errors.
	// It defaults to a No-Op warner implementation.
	Warner log.Warner
	// Metrics is the metrics interface to record metric data.
	// It defaults to a No-Op metrics implementation.
	Metrics metrics.DialMetrics
}

func (s *ServerSettings) setDefaults() {
	s.Resolver.setDefaults()

	if s.Port == 0 {
		const defaultPort = 53
		s.Port = defaultPort
	}

	if s.Filter == nil {
		s.Filter = filternoop.New()
	}

	if s.Logger == nil {
		s.Logger = lognoop.New()
	}

	if s.Metrics == nil {
		s.Metrics = metricsnoop.New()
	}

	if s.Cache == nil {
		// no-op metrics for no-op cache
		s.Cache = cachenoop.New(cachenoop.Settings{})
	}
}

func (s *ResolverSettings) setDefaults() {
	if len(s.DoTProviders) == 0 {
		s.DoTProviders = []provider.Provider{provider.Cloudflare()}
	}

	if s.Timeout == 0 {
		const defaultTimeout = 5 * time.Second
		s.Timeout = defaultTimeout
	}

	if s.Warner == nil {
		s.Warner = lognoop.New()
	}

	if s.Metrics == nil {
		s.Metrics = metricsnoop.New()
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
		subSection+"Listening port: "+fmt.Sprint(s.Port))

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
