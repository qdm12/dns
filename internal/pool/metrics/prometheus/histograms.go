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
			"Pool connections in use duration in seconds spent by address", []string{"address"},
			[]float64{
				time.Millisecond.Seconds(),
				10 * time.Millisecond.Seconds(),
				50 * time.Millisecond.Seconds(),
				100 * time.Millisecond.Seconds(),
				300 * time.Millisecond.Seconds(),
				500 * time.Millisecond.Seconds(),
				time.Second.Seconds(),
			}),
		lifetimes: helpers.NewHistogramVec(prefix, "pool_connection_lifetime",
			"Pool connection total lifetime duration in seconds by address", []string{"address"},
			[]float64{
				time.Second.Seconds(),
				10 * time.Second.Seconds(),
				20 * time.Second.Seconds(),
				30 * time.Second.Seconds(),
				time.Minute.Seconds(),
				2 * time.Minute.Seconds(),
				5 * time.Minute.Seconds(),
				15 * time.Minute.Seconds(),
				30 * time.Minute.Seconds(),
				time.Hour.Seconds(),
				2 * time.Hour.Seconds(),
			}),
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
