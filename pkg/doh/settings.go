package doh

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/qdm12/dns/internal/settings/defaults"
	"github.com/qdm12/dns/pkg/cache"
	cachenoop "github.com/qdm12/dns/pkg/cache/noop"
	"github.com/qdm12/dns/pkg/doh/metrics"
	metricsnoop "github.com/qdm12/dns/pkg/doh/metrics/noop"
	"github.com/qdm12/dns/pkg/filter"
	filternoop "github.com/qdm12/dns/pkg/filter/noop"
	"github.com/qdm12/dns/pkg/log"
	lognoop "github.com/qdm12/dns/pkg/log/noop"
	logmiddleware "github.com/qdm12/dns/pkg/middlewares/log"
	"github.com/qdm12/dns/pkg/provider"
	"github.com/qdm12/gotree"
	"github.com/qdm12/govalid/address"
	"github.com/qdm12/govalid/port"
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
	DoHProviders []string
	SelfDNS      SelfDNS
	Timeout      time.Duration
	// Warner is the warning logger to log dial errors.
	// It defaults to a No-Op warner implementation.
	Warner log.Warner
	// Metrics is the metrics interface to record metric data.
	// It defaults to a No-Op metrics implementation.
	Metrics metrics.DialMetrics
}

type SelfDNS struct {
	// for the internal HTTP client to resolve the DoH url hostname.
	DoTProviders []string
	DNSProviders []string
	Timeout      time.Duration
	IPv6         *bool
}

func (s *ServerSettings) SetDefaults() {
	s.Resolver.SetDefaults()
	s.LogMiddleware.SetDefaults()
	s.Address = defaults.String(s.Address, ":53")

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
	s.SelfDNS.SetDefaults()

	if len(s.DoHProviders) == 0 {
		s.DoHProviders = []string{"cloudflare"}
	}

	const defaultTimeout = 5 * time.Second
	s.Timeout = defaults.Duration(s.Timeout, defaultTimeout)

	if s.Warner == nil {
		s.Warner = lognoop.New()
	}

	if s.Metrics == nil {
		s.Metrics = metricsnoop.New()
	}
}

func (s *SelfDNS) SetDefaults() {
	const defaultTimeout = 5 * time.Second
	s.Timeout = defaults.Duration(s.Timeout, defaultTimeout)

	if len(s.DoTProviders) == 0 {
		s.DoTProviders = []string{"cloudflare"}
	}
	// No default DNS fallback server for the internal HTTP client
	// to avoid leaking we are using a DoH server.

	s.IPv6 = defaults.BoolPtr(s.IPv6, false)
}

var (
	ErrListeningAddressNotValid = errors.New("listening address is not valid")
)

func (s ServerSettings) Validate() (err error) {
	err = s.Resolver.Validate()
	if err != nil {
		return fmt.Errorf("failed validating resolver settings: %w", err)
	}

	const defaultUDPPort = 53
	_, err = address.Validate(s.Address,
		address.OptionListening(
			os.Getuid(), port.OptionListeningPortPrivilegedAllowed(defaultUDPPort)))
	if err != nil {
		return fmt.Errorf("%w: %s", ErrListeningAddressNotValid, s.Address)
	}

	err = s.LogMiddleware.Validate()
	if err != nil {
		return fmt.Errorf("failed validating log middleware settings: %w", err)
	}

	return nil
}

func (s ResolverSettings) Validate() (err error) {
	for _, s := range s.DoHProviders {
		_, err = provider.Parse(s)
		if err != nil {
			return fmt.Errorf("invalid DoH provider: %w", err)
		}
	}

	err = s.SelfDNS.Validate()
	if err != nil {
		return fmt.Errorf("failed validating DoH self DNS settings: %w", err)
	}

	return nil
}

func (s SelfDNS) Validate() (err error) {
	for _, s := range s.DoTProviders {
		_, err = provider.Parse(s)
		if err != nil {
			return fmt.Errorf("invalid DoT provider: %w", err)
		}
	}

	for _, s := range s.DNSProviders {
		_, err = provider.Parse(s)
		if err != nil {
			return fmt.Errorf("invalid fallback plaintext DNS provider: %w", err)
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

func (s *SelfDNS) String() string {
	return s.ToLinesNode().String()
}

func (s *ServerSettings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DoH server settings:")
	node.Appendf("Listening address: %s", s.Address)
	node.AppendNode(s.Resolver.ToLinesNode())
	return node
}

func (s *ResolverSettings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DoH resolver settings:")

	DoTProvidersNode := node.Appendf("DNS over HTTPs providers:")
	for _, provider := range s.DoHProviders {
		DoTProvidersNode.Appendf(strings.Title(provider))
	}

	node.AppendNode(s.SelfDNS.ToLinesNode())

	node.Appendf("Query timeout: %s", s.Timeout)

	return node
}

func (s *SelfDNS) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Internal DNS settings:")
	node.Appendf("Query timeout: %s", s.Timeout)

	connectOver := "IPv4"
	if *s.IPv6 {
		connectOver = "IPv6"
	}
	node.Appendf("Connecting over: %s", connectOver)

	if len(s.DoTProviders) > 0 {
		DoTProvidersNode := node.Appendf("DNS over TLS providers:")
		for _, provider := range s.DoTProviders {
			DoTProvidersNode.Appendf(strings.Title(provider))
		}
	}

	if len(s.DNSProviders) > 0 {
		fallbackPlaintextProvidersNode := node.Appendf("Fallback plaintext DNS providers:")
		for _, provider := range s.DNSProviders {
			fallbackPlaintextProvidersNode.Appendf(strings.Title(provider))
		}
	}

	return node
}
