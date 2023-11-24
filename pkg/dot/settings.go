package dot

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/qdm12/dns/v2/internal/picker"
	metricsnoop "github.com/qdm12/dns/v2/pkg/dot/metrics/noop"
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
	ListeningAddress *string
	// Middlewares is a list of middlewares to use.
	// The first one is the first wrapper, and the last one
	// is the last wrapper of the handlers in the chain.
	Middlewares []Middleware
	// Logger is the logger to log information.
	// It defaults to a No-Op logger implementation.
	Logger Logger
	// Metrics is the metrics interface to record metric data.
	// It defaults to a No-Op metrics implementation.
	// Metrics metrics.DialMetrics
}

type ResolverSettings struct {
	DoTProviders []provider.Provider
	Timeout      time.Duration
	// IPVersion defines the only IP version to use to connect to
	// upstream DNS over TLS servers. If left unset, it defaults to
	// "ipv4".
	IPVersion string
	// Warner is the warning logger to log dial errors.
	// It defaults to a No-Op warner implementation.
	Warner Warner
	// Metrics is the metrics interface to record metric data.
	// It defaults to a No-Op metrics implementation.
	Metrics Metrics
	// Picker is the picker to use for each upstream call to pick
	// a server and/or IP address. It must be thread safe.
	// It defaults to a fast thread safe pseudo random picker
	// with uniform distribution.
	Picker Picker
}

func (s *ServerSettings) SetDefaults() {
	s.Resolver.SetDefaults()
	s.ListeningAddress = gosettings.DefaultPointer(s.ListeningAddress, ":53")
	s.Logger = gosettings.DefaultComparable[Logger](s.Logger, lognoop.New())
}

func (s *ResolverSettings) SetDefaults() {
	s.DoTProviders = gosettings.DefaultSlice(s.DoTProviders,
		[]provider.Provider{provider.Cloudflare()})
	// No default DNS fallback server for the internal HTTP client
	// to avoid leaking we are using a DoT server.
	const defaultTimeout = 5 * time.Second
	s.Timeout = gosettings.DefaultComparable(s.Timeout, defaultTimeout)
	s.IPVersion = gosettings.DefaultComparable(s.IPVersion, "ipv4")
	s.Warner = gosettings.DefaultComparable[Warner](s.Warner, lognoop.New())
	s.Metrics = gosettings.DefaultComparable[Metrics](s.Metrics, metricsnoop.New())
	s.Picker = gosettings.DefaultComparable[Picker](s.Picker, picker.New())
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
	err = validate.ListeningAddress(*s.ListeningAddress, os.Getuid(), defaultUDPPort)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrListeningAddressNotValid, *s.ListeningAddress)
	}

	return nil
}

var (
	ErrDoTProvidersNotSet = errors.New("DoT providers are not set")
)

func (s ResolverSettings) Validate() (err error) {
	if len(s.DoTProviders) == 0 {
		// just in case the user sets the slice to the empty non-nil slice
		return fmt.Errorf("%w", ErrDoTProvidersNotSet)
	}

	for _, provider := range s.DoTProviders {
		err = provider.ValidateForDoT()
		if err != nil {
			return fmt.Errorf("DNS over TLS provider %s: %w", provider.Name, err)
		}
	}

	err = validate.IsOneOf(s.IPVersion, "ipv4", "ipv6")
	if err != nil {
		return fmt.Errorf("IP version: %w", err)
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
	node.Appendf("Listening address: %s", *s.ListeningAddress)
	node.AppendNode(s.Resolver.ToLinesNode())
	return node
}

func (s *ResolverSettings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DoT resolver settings:")

	DoTProvidersNode := node.Appendf("DNS over TLS providers:")
	caser := cases.Title(language.English)
	for _, provider := range s.DoTProviders {
		DoTProvidersNode.Appendf(caser.String(provider.Name))
	}

	node.Appendf("Query timeout: %s", s.Timeout)
	node.Appendf("Connecting over: %s", s.IPVersion)

	return node
}
