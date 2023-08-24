package prometheus

import "github.com/prometheus/client_golang/prometheus"

type Registry interface {
	prometheus.Registerer
}
