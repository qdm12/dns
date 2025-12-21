package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type histograms struct {
	useTimes  *prometheus.HistogramVec
	lifetimes *prometheus.HistogramVec
}

func newHistograms(settings prom.Settings) (h *histograms, err error) {
	prefix := settings.Prefix
	h = &histograms{
		useTimes: helpers.NewHistogramVec(prefix, "pool_connection_usetime",
			"Pool connections in use duration spent by address", []string{"address"}, nil),
		lifetimes: helpers.NewHistogramVec(prefix, "pool_connection_lifetime",
			"Pool connection total lifetime duration in seconds by address", []string{"address"}, nil),
	}
	err = helpers.Register(settings.Registry, h.useTimes, h.lifetimes)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (h *histograms) RecordUseTime(address string, duration time.Duration) {
	h.useTimes.WithLabelValues(address).Observe(duration.Seconds())
}

func (h *histograms) RecordLifetime(address string, duration time.Duration) {
	h.lifetimes.WithLabelValues(address).Observe(duration.Seconds())
}
