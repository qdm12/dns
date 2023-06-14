package dot

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/qdm12/dns/v2/internal/picker"
	"github.com/qdm12/dns/v2/pkg/cache"
	cachenoop "github.com/qdm12/dns/v2/pkg/cache/noop"
	"github.com/qdm12/dns/v2/pkg/dot/metrics"
	metricsnoop "github.com/qdm12/dns/v2/pkg/dot/metrics/noop"
	"github.com/qdm12/dns/v2/pkg/filter"
	filternoop "github.com/qdm12/dns/v2/pkg/filter/noop"
	"github.com/qdm12/dns/v2/pkg/log"
	lognoop "github.com/qdm12/dns/v2/pkg/log/noop"
	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ServerSettings struct {
	Resolver         ResolverSettings
	ListeningAddress string
	// Middlewares is a list of middlewares to use.
	// The first one is the first wrapper, and the last one
	// is the last wrapper of the handlers in the chain.
	Middlewares []Middleware
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
	// Metrics metrics.DialMetrics
}

type ResolverSettings struct {
	DoTProviders []string
	DNSProviders []string
	Timeout      time.Duration
	// IPv6 is false if the resolver should connect to
	// nameservers over IPv4 and true to connect over IPv6.
	IPv6 bool
	// Warner is the warning logger to log dial errors.
	// It defaults to a No-Op warner implementation.
	Warner log.Warner
	// Metrics is the metrics interface to record metric data.
	// It defaults to a No-Op metrics implementation.
	Metrics metrics.DialMetrics
	// Picker is the picker to use for each upstream call to pick
	// a server and/or IP address. It must be thread safe.
	// It defaults to a fast thread safe pseudo random picker
	// with uniform distribution.
	Picker Picker
}

func (s *ServerSettings) SetDefaults() {
	s.Resolver.SetDefaults()
	s.ListeningAddress = gosettings.DefaultString(s.ListeningAddress, ":53")
	s.Filter = gosettings.DefaultInterface(s.Filter, filternoop.New())
	s.Logger = gosettings.DefaultInterface(s.Logger, lognoop.New())
	s.Cache = gosettings.DefaultInterface(s.Cache, cachenoop.New(cachenoop.Settings{}))
}

func (s *ResolverSettings) SetDefaults() {
	s.DoTProviders = gosettings.DefaultSlice(s.DoTProviders, []string{"cloudflare"})
	const defaultTimeout = 5 * time.Second
	s.Timeout = gosettings.DefaultNumber(s.Timeout, defaultTimeout)
	s.Warner = gosettings.DefaultInterface(s.Warner, lognoop.New())
	s.Metrics = gosettings.DefaultInterface(s.Metrics, metricsnoop.New())
	s.Picker = gosettings.DefaultInterface(s.Picker, picker.New())
}

var (
	ErrListeningAddressNotValid = errors.New("listening address is not valid")
)

func (s ServerSettings) Validate() (err error) {
	err = s.Resolver.Validate()
	if err != nil {
		return fmt.Errorf("resolver settings: %w", err)
	}

	const defaultUDPPort = 53
	err = validate.ListeningAddress(s.ListeningAddress, os.Getuid(), defaultUDPPort)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrListeningAddressNotValid, s.ListeningAddress)
	}

	return nil
}

func (s ResolverSettings) Validate() (err error) {
	for _, s := range s.DoTProviders {
		_, err = provider.Parse(s)
		if err != nil {
			return fmt.Errorf("DoT provider: %w", err)
		}
	}

	for _, s := range s.DNSProviders {
		_, err = provider.Parse(s)
		if err != nil {
			return fmt.Errorf("plaintext DNS provider: %w", err)
		}
	}

	return nil
}

func (s *ServerSettings) String() string {
	return s.ToLinesNode().String()
}

func (s *ResolverSettings) String() string {
	return s.ToLinesNode().String()
}

func (s *ServerSettings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DoT server settings:")
	node.Appendf("Listening address: %s", s.ListeningAddress)
	node.AppendNode(s.Resolver.ToLinesNode())
	return node
}

func (s *ResolverSettings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DoT resolver settings:")

	DoTProvidersNode := node.Appendf("DNS over TLS providers:")
	caser := cases.Title(language.English)
	for _, provider := range s.DoTProviders {
		DoTProvidersNode.Appendf(caser.String(provider))
	}

	fallbackPlaintextProvidersNode := node.Appendf("Fallback plaintext DNS providers:")
	for _, provider := range s.DNSProviders {
		fallbackPlaintextProvidersNode.Appendf(caser.String(provider))
	}

	node.Appendf("Quey timeout: %s", s.Timeout)

	connectOver := "IPv4"
	if s.IPv6 {
		connectOver = "IPv6"
	}
	node.Appendf("Connecting over: %s", connectOver)

	return node
}
