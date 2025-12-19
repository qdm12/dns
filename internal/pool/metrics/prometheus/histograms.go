package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type histograms struct {
	useTimes *prometheus.HistogramVec
}

func newHistograms(settings prom.Settings) (h *histograms, err error) {
	prefix := settings.Prefix
	h = &histograms{
		useTimes: helpers.NewHistogramVec(prefix, "pool_use_time",
			"Pool connections in use duration spent by address", []string{"address"}, nil),
	}
	err = helpers.Register(settings.Registry, h.useTimes)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (h *histograms) RecordUseTime(address string, duration time.Duration) {
	h.useTimes.WithLabelValues(address).Observe(duration.Seconds())
}
