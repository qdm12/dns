package doh

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/qdm12/dns/v2/internal/picker"
	metricsnoop "github.com/qdm12/dns/v2/pkg/doh/metrics/noop"
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
	// Logger is the logger to log information.
	// It defaults to a No-Op logger implementation.
	Logger Logger
}

type ResolverSettings struct {
	DoHProviders []string
	SelfDNS      SelfDNS
	Timeout      time.Duration
	// Warner is the warning logger to log dial errors.
	// It defaults to a No-Op warner implementation.
	Warner Warner
	// Metrics is the metrics interface to record metric data.
	// It defaults to a No-Op metrics implementation.
	Metrics Metrics
	// Picker is the picker to use for each upstream call to pick
	// a server from a pool of servers. It must be thread safe.
	// It defaults to a fast thread safe pseudo random picker
	// with uniform distribution.
	Picker Picker
}

type SelfDNS struct {
	// for the internal HTTP client to resolve the DoH url hostname.
	DoTProviders []string
	DNSProviders []string
	Timeout      time.Duration
	IPv6         bool
}

func (s *ServerSettings) SetDefaults() {
	s.Resolver.SetDefaults()
	s.ListeningAddress = gosettings.DefaultString(s.ListeningAddress, ":53")
	s.Logger = gosettings.DefaultInterface(s.Logger, lognoop.New())
}

func (s *ResolverSettings) SetDefaults() {
	s.SelfDNS.SetDefaults()
	s.DoHProviders = gosettings.DefaultSlice(s.DoHProviders, []string{"cloudflare"})
	const defaultTimeout = 5 * time.Second
	s.Timeout = gosettings.DefaultNumber(s.Timeout, defaultTimeout)
	s.Warner = gosettings.DefaultInterface(s.Warner, lognoop.New())
	s.Metrics = gosettings.DefaultInterface(s.Metrics, metricsnoop.New())
	s.Picker = gosettings.DefaultInterface(s.Picker, picker.New())
}

func (s *SelfDNS) SetDefaults() {
	const defaultTimeout = 5 * time.Second
	s.Timeout = gosettings.DefaultNumber(s.Timeout, defaultTimeout)
	s.DoTProviders = gosettings.DefaultSlice(s.DoTProviders, []string{"cloudflare"})
	// No default DNS fallback server for the internal HTTP client
	// to avoid leaking we are using a DoH server.
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
	for _, s := range s.DoHProviders {
		_, err = provider.Parse(s)
		if err != nil {
			return fmt.Errorf("DoH provider: %w", err)
		}
	}

	err = s.SelfDNS.Validate()
	if err != nil {
		return fmt.Errorf("DoH self DNS settings: %w", err)
	}

	return nil
}

func (s SelfDNS) Validate() (err error) {
	for _, s := range s.DoTProviders {
		_, err = provider.Parse(s)
		if err != nil {
			return fmt.Errorf("DoT provider: %w", err)
		}
	}

	for _, s := range s.DNSProviders {
		_, err = provider.Parse(s)
		if err != nil {
			return fmt.Errorf("fallback plaintext DNS provider: %w", err)
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
	node.Appendf("Listening address: %s", s.ListeningAddress)
	node.AppendNode(s.Resolver.ToLinesNode())
	return node
}

func (s *ResolverSettings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DoH resolver settings:")

	DoTProvidersNode := node.Appendf("DNS over HTTPs providers:")
	caser := cases.Title(language.English)
	for _, provider := range s.DoHProviders {
		DoTProvidersNode.Appendf(caser.String(provider))
	}

	node.AppendNode(s.SelfDNS.ToLinesNode())

	node.Appendf("Query timeout: %s", s.Timeout)

	return node
}

func (s *SelfDNS) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Internal DNS settings:")
	node.Appendf("Query timeout: %s", s.Timeout)

	connectOver := "IPv4"
	if s.IPv6 {
		connectOver = "IPv6"
	}
	node.Appendf("Connecting over: %s", connectOver)

	caser := cases.Title(language.English)

	if len(s.DoTProviders) > 0 {
		DoTProvidersNode := node.Appendf("DNS over TLS providers:")
		for _, provider := range s.DoTProviders {
			DoTProvidersNode.Appendf(caser.String(provider))
		}
	}

	if len(s.DNSProviders) > 0 {
		fallbackPlaintextProvidersNode := node.Appendf("Fallback plaintext DNS providers:")
		for _, provider := range s.DNSProviders {
			fallbackPlaintextProvidersNode.Appendf(caser.String(provider))
		}
	}

	return node
}
