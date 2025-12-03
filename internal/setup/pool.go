package setup

import (
	"fmt"

	noopmetrics "github.com/qdm12/dns/v2/internal/pool/metrics/noop"
	prometheusmetrics "github.com/qdm12/dns/v2/internal/pool/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

func poolMetrics(metricsType string, //nolint:ireturn
	commonPrometheus prometheus.Settings) (
	metrics PoolMetrics, err error,
) {
	switch metricsType {
	case noopString:
		return noopmetrics.New(), nil
	case prometheusString:
		prometheusSettings := prometheusmetrics.Settings{
			Prometheus: commonPrometheus,
		}
		return prometheusmetrics.New(prometheusSettings)
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", metricsType))
	}
}
