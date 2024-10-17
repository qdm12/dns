package setup

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config"
	promcommon "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	noopmetrics "github.com/qdm12/dns/v2/pkg/middlewares/filter/metrics/noop"
	prommetrics "github.com/qdm12/dns/v2/pkg/middlewares/filter/metrics/prometheus"
)

func BuildFilterMetrics(userSettings config.Metrics, //nolint:ireturn
	registry PrometheusRegistry) (
	metrics FilterMetrics, err error,
) {
	switch userSettings.Type {
	case noopString:
		return noopmetrics.New(), nil
	case prometheusString:
		settings := prommetrics.Settings{
			Prometheus: promcommon.Settings{
				Registry: registry,
				Prefix:   *userSettings.Prometheus.Subsystem,
			},
		}
		metrics, err = prommetrics.New(settings)
		if err != nil {
			return nil, fmt.Errorf("setting up Prometheus metrics: %w", err)
		}
		return metrics, nil
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", userSettings.Type))
	}
}
