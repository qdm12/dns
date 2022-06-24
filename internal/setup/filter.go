package setup

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/pkg/filter/metrics"
	noopmetrics "github.com/qdm12/dns/v2/pkg/filter/metrics/noop"
	prommetrics "github.com/qdm12/dns/v2/pkg/filter/metrics/prometheus"
	promcommon "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

func FilterMetrics(userSettings settings.Metrics, //nolint:ireturn
	registry prometheus.Registerer) (
	metrics metrics.Interface, err error) {
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
			return nil, fmt.Errorf("setup Prometheus metrics: %w", err)
		}
		return metrics, nil
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", userSettings.Type))
	}
}