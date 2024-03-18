package dnssec

import (
	"github.com/qdm12/dns/v2/pkg/log/noop"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// Logger is the logger to use.
	// It defaults to a No-op implementation.
	Logger Logger
}

func (s *Settings) SetDefaults() {
	s.Logger = gosettings.DefaultComparable[Logger](s.Logger, noop.New())
}

func (s *Settings) Validate() error { return nil }

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DNSSEC settings:")
	return node
}
