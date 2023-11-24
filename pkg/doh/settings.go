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
	ListeningAddress *string
	// Middlewares is a list of middlewares to use.
	// The first one is the first wrapper, and the last one
	// is the last wrapper of the handlers in the chain.
	Middlewares []Middleware
	// Logger is the logger to log information.
	// It defaults to a No-Op logger implementation.
	Logger Logger
}

type ResolverSettings struct {
	// UpstreamResolvers is a list of DNS over TLS upstream resolvers
	// to use.
	UpstreamResolvers []provider.Provider
	// IPVersion indicates whether to use IPv4 only or IPv6 only for
	// DNS over HTTPS. The hardcoded resolver used by the DoH HTTP
	// client will return only IP addresses matching the version set
	// from all the providers. If left unset, it defaults to "ipv4".
	IPVersion string
	Timeout   time.Duration
	// Metrics is the metrics interface to record metric data.
	// It defaults to a No-Op metrics implementation.
	Metrics Metrics
	// Picker is the picker to use for each upstream call to pick
	// a server from a pool of servers. It must be thread safe.
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
	s.UpstreamResolvers = gosettings.DefaultSlice(s.UpstreamResolvers,
		[]provider.Provider{provider.Cloudflare()})
	s.IPVersion = gosettings.DefaultComparable(s.IPVersion, "ipv4")
	const defaultTimeout = 5 * time.Second
	s.Timeout = gosettings.DefaultComparable(s.Timeout, defaultTimeout)
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
	ErrUpstreamResolversNotSet = errors.New("upstream resolvers not set")
)

func (s ResolverSettings) Validate() (err error) {
	if len(s.UpstreamResolvers) == 0 {
		// just in case the user sets the slice to the empty non-nil slice
		return fmt.Errorf("%w", ErrUpstreamResolversNotSet)
	}

	err = validate.IsOneOf(s.IPVersion, "ipv4", "ipv6")
	if err != nil {
		return fmt.Errorf("IP version: %w", err)
	}

	for _, upstreamResolver := range s.UpstreamResolvers {
		err = upstreamResolver.ValidateForDoH(s.IPVersion == "ipv6")
		if err != nil {
			return fmt.Errorf("upstream resolver %s: %w", upstreamResolver.Name, err)
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
	node = gotree.New("DoH server settings:")
	node.Appendf("Listening address: %s", *s.ListeningAddress)
	node.AppendNode(s.Resolver.ToLinesNode())
	return node
}

func (s *ResolverSettings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DoH resolver settings:")

	upstreamResolversNode := node.Appendf("Upstream resolvers:")
	caser := cases.Title(language.English)
	for _, provider := range s.UpstreamResolvers {
		upstreamResolversNode.Appendf(caser.String(provider.Name))
	}

	node.Appendf("Connecting over %s", s.IPVersion)
	node.Appendf("Query timeout: %s", s.Timeout)

	return node
}
