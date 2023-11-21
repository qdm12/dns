package noop

import (
	"github.com/qdm12/dns/v2/pkg/middlewares/cache/metrics/noop"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// Metrics is the metrics interface to record the cache type.
	// It defaults to a No-Op metric implementation.
	Metrics Metrics
}

func (s *Settings) SetDefaults() {
	s.Metrics = gosettings.DefaultComparable[Metrics](s.Metrics, noop.New())
}

func (s Settings) Validate() (err error) {
	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	return gotree.New("Noop cache settings:")
}
