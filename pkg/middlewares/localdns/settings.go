package localdns

import (
	"net/netip"
	"strings"

	"github.com/qdm12/dns/v2/pkg/log/noop"
	"github.com/qdm12/dns/v2/pkg/nameserver"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// Resolvers is the list of resolvers to use to resolve the
	// local domain names. They are each tried after the other
	// in order, until one returns an answer for the question.
	// If left empty, the local nameservers found are used.
	Resolvers []netip.AddrPort
	// PublicNamesAsLocal is a list of local names that should be considered local.
	PublicNamesAsLocal []string
	// Logger is the logger to use.
	// It defaults to a No-op implementation.
	Logger Logger
	// TimeoutWarn indicates whether to log timeout errors at the
	// warning level or at the debug level. It defaults to false.
	TimeoutWarn *bool
}

func (s *Settings) SetDefaults() {
	privateNameservers, _ := nameserver.GetPrivateDNSServers()
	s.Resolvers = gosettings.DefaultSlice(s.Resolvers, addrsToAddr53(privateNameservers))
	s.PublicNamesAsLocal = gosettings.DefaultSlice(s.PublicNamesAsLocal, []string{})
	s.Logger = gosettings.DefaultComparable[Logger](s.Logger, noop.New())
	s.TimeoutWarn = gosettings.DefaultPointer(s.TimeoutWarn, false)
}

func addrsToAddr53(addrs []netip.Addr) (addrPorts []netip.AddrPort) {
	addrPorts = make([]netip.AddrPort, len(addrs))
	const dnsPort = 53
	for i := range addrs {
		addrPorts[i] = netip.AddrPortFrom(addrs[i], dnsPort)
	}
	return addrPorts
}

func (s *Settings) Validate() (err error) {
	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Local forwarding middleware settings:")

	node.Appendf("Log timeout errors at the warning level: %s", gosettings.BoolToYesNo(s.TimeoutWarn))

	resolversNode := node.Appendf("Local resolvers:")
	for _, resolver := range s.Resolvers {
		resolversNode.Appendf("%s", resolver)
	}

	if len(s.PublicNamesAsLocal) > 0 {
		node.Appendf("Public names considered local: %s",
			strings.Join(s.PublicNamesAsLocal, ", "))
	}

	return node
}
