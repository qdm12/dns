package doh

import (
	"errors"
	"fmt"
	"time"

	metricsnoop "github.com/qdm12/dns/v2/pkg/doh/metrics/noop"
	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Settings struct {
	// UpstreamResolvers is a list of DNS over TLS upstream resolvers
	// to use.
	UpstreamResolvers []provider.Provider
	// Timeout is the maximum duration to wait for a response from
	// upstream DNS over HTTPS servers. If left unset, it defaults to
	// 5 seconds.
	Timeout time.Duration
	// IPVersion indicates whether to use IPv4 only or IPv6 only for
	// DNS over HTTPS. The hardcoded resolver used by the DoH HTTP
	// client will return only IP addresses matching the version set
	// from all the providers. If left unset, it defaults to "ipv4".
	IPVersion string
	// Metrics is the metrics interface to record metric data.
	// It defaults to a No-Op metrics implementation.
	Metrics Metrics
}

func (s *Settings) SetDefaults() {
	s.UpstreamResolvers = gosettings.DefaultSlice(s.UpstreamResolvers,
		[]provider.Provider{provider.Cloudflare()})
	const defaultTimeout = 5 * time.Second
	s.Timeout = gosettings.DefaultComparable(s.Timeout, defaultTimeout)
	s.IPVersion = gosettings.DefaultComparable(s.IPVersion, "ipv4")
	s.Metrics = gosettings.DefaultComparable[Metrics](s.Metrics, metricsnoop.New())
}

var ErrUpstreamResolversNotSet = errors.New("upstream resolvers not set")

func (s Settings) Validate() (err error) {
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

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DoH resolver settings:")

	upstreamResolversNode := node.Append("Upstream resolvers:")
	caser := cases.Title(language.English)
	for _, provider := range s.UpstreamResolvers {
		upstreamResolversNode.Append(caser.String(provider.Name))
	}

	node.Appendf("Query timeout: %s", s.Timeout)

	connectingOver := "ipv4"
	if s.IPVersion == "ipv6" {
		connectingOver += " and ipv6"
	}
	node.Appendf("Connecting over: %s", connectingOver)

	return node
}
