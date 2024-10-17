package setup

import (
	"fmt"

	"github.com/qdm12/dns/v2/pkg/metrics/prometheus"
	metricsmiddleware "github.com/qdm12/dns/v2/pkg/middlewares/metrics"
	noopmetrics "github.com/qdm12/dns/v2/pkg/middlewares/metrics/noop"
	prometheusmetrics "github.com/qdm12/dns/v2/pkg/middlewares/metrics/prometheus"
)

func middlewareMetrics(metricsType string,
	commonPrometheus prometheus.Settings) (
	middleware *metricsmiddleware.Middleware, err error,
) {
	var metrics interface {
		RequestsInc()
		QuestionsInc(class, qType string)
		RcodeInc(rcode string)
		AnswersInc(class, qType string)
		ResponsesInc()
		InFlightRequestsInc()
		InFlightRequestsDec()
	}
	switch metricsType {
	case noopString:
		metrics = noopmetrics.New()
	case prometheusString:
		promSettings := prometheusmetrics.Settings{
			Prometheus: commonPrometheus,
		}
		metrics, err = prometheusmetrics.New(promSettings)
		if err != nil {
			return nil, fmt.Errorf("prometheus metrics: %w", err)
		}
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", metricsType))
	}

	settings := metricsmiddleware.Settings{
		Metrics: metrics,
	}
	return metricsmiddleware.New(settings)
}
