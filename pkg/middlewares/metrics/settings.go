package metrics

import (
	"github.com/qdm12/dns/v2/pkg/middlewares/metrics/noop"
	"github.com/qdm12/gosettings"
)

type Settings struct {
	// Metrics is the interface to interact with metrics in the
	// DNS middleware. It defaults to a No-Op implementation.
	Metrics Metrics
}

func (s *Settings) SetDefaults() {
	s.Metrics = gosettings.DefaultInterface(s.Metrics, noop.New())
}

func (s Settings) Validate() (err error) {
	return nil
}
