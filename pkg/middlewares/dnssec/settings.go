package dnssec

import (
	"time"

	"github.com/qdm12/dns/v2/pkg/log/noop"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// Logger is the logger to use.
	// It defaults to a No-op implementation.
	Logger Logger
	// RootTrustAnchorRefreshPeriod determines how often the middleware should
	// query the root DNSKEY RRSet and refresh the local root trust anchors.
	// Set it to 0 to disable periodic refreshes.
	RootTrustAnchorRefreshPeriod *time.Duration
}

func (s *Settings) SetDefaults() {
	s.Logger = gosettings.DefaultComparable[Logger](s.Logger, noop.New())
	const defaultRefreshPeriod = 7 * 24 * time.Hour
	s.RootTrustAnchorRefreshPeriod = gosettings.DefaultPointer(
		s.RootTrustAnchorRefreshPeriod, defaultRefreshPeriod)
}

func (s *Settings) Validate() error { return nil }

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DNSSEC settings:")
	if *s.RootTrustAnchorRefreshPeriod == 0 {
		node.Appendf("Root trust anchor refresh: disabled")
	} else {
		node.Appendf("Root trust anchor refresh: every %s",
			*s.RootTrustAnchorRefreshPeriod)
	}
	return node
}
