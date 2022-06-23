package setup

import (
	"fmt"

	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	noopmetrics "github.com/qdm12/dns/v2/pkg/middlewares/metrics/noop"
	prometheusmetrics "github.com/qdm12/dns/v2/pkg/middlewares/metrics/prometheus"
)

func middlewareMetrics(metricsType string, //nolint:ireturn
	commonPrometheus prometheus.Settings) (
	metrics MiddlewareMetrics, err error) {
	switch metricsType {
	case noopString:
		return noopmetrics.New(), nil
	case prometheusString:
		settings := prometheusmetrics.Settings{
			Prometheus: commonPrometheus,
		}
		return prometheusmetrics.New(settings)
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", metricsType))
	}
}
