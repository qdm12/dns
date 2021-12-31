package metrics

import "github.com/qdm12/dns/pkg/middlewares/metrics/noop"

type Settings struct {
	// Metrics is the interface to interact with metrics in the
	// DNS middleware. It defaults to a No-Op implementation.
	Metrics Interface
}

func (s *Settings) SetDefaults() {
	if s.Metrics == nil {
		s.Metrics = noop.New()
	}
}
