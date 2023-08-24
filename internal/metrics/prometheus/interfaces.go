package prometheus

import (
	dto "github.com/prometheus/client_model/go"
)

type Gatherer interface {
	Gather() ([]*dto.MetricFamily, error)
}
