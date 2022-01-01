package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/pkg/metrics/prometheus"
)

type gauges struct {
	requestsInFlight prometheus.Gauge
}

func newGauges(settings prom.Settings) (g *gauges, err error) {
	prefix := *settings.Prefix
	g = &gauges{
		requestsInFlight: helpers.NewGauge(prefix, "requests_inflight",
			"Requests in flight in the server"),
	}

	err = helpers.Register(settings.Registry, g.requestsInFlight)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (g *gauges) InFlightRequestsInc() {
	g.requestsInFlight.Inc()
}

func (g *gauges) InFlightRequestsDec() {
	g.requestsInFlight.Dec()
}
