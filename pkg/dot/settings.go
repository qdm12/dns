package dot

import (
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
	"github.com/qdm12/gotree"
)

type ServerSettings struct {
	Resolver      ResolverSettings
	Address       string
	LogMiddleware logmiddleware.Settings
	// Cache is the cache to use in the server.
	// It defaults to a No-Op cache implementation with
	// a No-Op cache metrics implementation.
	Cache cache.Interface
	// Filter is the filter for DNS requests and responses.
	// It defaults to a No-Op filter implementation.
	Filter filter.Interface
	// Logger is the logger to log information.
	// It defaults to a No-Op logger implementation.
	Logger log.Logger
	// Metrics is the metrics interface to record metric data.
	// It defaults to a No-Op metrics implementation.
	Metrics metrics.Interface
}

type ResolverSettings struct {
	DoTProviders []string
	DNSProviders []string
	Timeout      time.Duration
	// IPv6 is false if the resolver should connect to
	// nameservers over IPv4 and true to connect over IPv6.
	// It cannot be nil in the internal state.
	IPv6 *bool
	// Warner is the warning logger to log dial errors.
	// It defaults to a No-Op warner implementation.
	Warner log.Warner
	// Metrics is the metrics interface to record metric data.
	// It defaults to a No-Op metrics implementation.
	Metrics metrics.DialMetrics
}

func (s *ServerSettings) SetDefaults() {
	s.Resolver.SetDefaults()

	if s.Address == "" {
		const defaultAddress = ":53"
		s.Address = defaultAddress
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

func (s *ResolverSettings) SetDefaults() {
	if len(s.DoTProviders) == 0 {
		s.DoTProviders = []string{"cloudflare"}
	}

	if s.Timeout == 0 {
		const defaultTimeout = 5 * time.Second
		s.Timeout = defaultTimeout
	}

	if s.IPv6 == nil {
		ipv6 := false
		s.IPv6 = &ipv6
	}

	if s.Warner == nil {
		s.Warner = lognoop.New()
	}

	if s.Metrics == nil {
		s.Metrics = metricsnoop.New()
	}
}

func (s *ServerSettings) String() string {
	return s.ToLinesNode().String()
}

func (s *ResolverSettings) String() string {
	return s.ToLinesNode().String()
}

func (s *ServerSettings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DoT server settings:")
	node.Appendf("Listening address: %s", s.Address)
	node.AppendNode(s.Resolver.ToLinesNode())
	return node
}

func (s *ResolverSettings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DoT resolver settings:")

	DoTProvidersNode := node.Appendf("DNS over TLS providers:")
	for _, provider := range s.DoTProviders {
		DoTProvidersNode.Appendf(strings.Title(provider))
	}

	fallbackPlaintextProvidersNode := node.Appendf("Fallback plaintext DNS providers:")
	for _, provider := range s.DNSProviders {
		fallbackPlaintextProvidersNode.Appendf(strings.Title(provider))
	}

	node.Appendf("Quey timeout: %s", s.Timeout)

	connectOver := "IPv4"
	if *s.IPv6 {
		connectOver = "IPv6"
	}
	node.Appendf("Connecting over: %s", connectOver)

	return node
}
