package localdns

import (
	"errors"
	"fmt"
	"net/netip"

	"github.com/qdm12/dns/v2/pkg/log/noop"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// Resolvers is the list of resolvers to use to resolve the
	// local domain names. They are each tried after the other
	// in order, until one returns an answer for the question.
	// This field must be set and non empty.
	Resolvers []netip.AddrPort
	// Logger is the logger to use.
	// It defaults to a No-op implementation.
	Logger Logger
}

func (s *Settings) SetDefaults() {
	s.Logger = gosettings.DefaultComparable[Logger](s.Logger, noop.New())
}

var (
	ErrResolversNotSet = errors.New("resolvers are not set")
)

func (s *Settings) Validate() (err error) {
	if len(s.Resolvers) == 0 {
		return fmt.Errorf("%w", ErrResolversNotSet)
	}

	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Local forwarding middleware settings:")

	resolversNode := gotree.New("Local resolvers:")
	for _, resolver := range s.Resolvers {
		resolversNode.Appendf("%s", resolver)
	}
	node.AppendNode(resolversNode)

	return node
}
