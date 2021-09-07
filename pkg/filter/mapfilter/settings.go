package mapfilter

import (
	"reflect"
	"strings"

	"github.com/qdm12/dns/pkg/filter/metrics"
	"github.com/qdm12/dns/pkg/filter/metrics/noop"
	"github.com/qdm12/dns/pkg/filter/metrics/prometheus"
	"github.com/qdm12/dns/pkg/filter/update"
)

type Settings struct {
	Update  update.Settings
	Metrics metrics.Interface
}

func (s *Settings) setDefaults() {
	if s.Metrics == nil {
		s.Metrics = noop.New()
	}
}

func (s *Settings) String() string {
	const (
		subSection = " |--"
		indent     = "    " // used if lines already contain the subSection
	)
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *Settings) Lines(indent, subSection string) (lines []string) {
	lines = append(lines, s.Update.Lines(indent, subSection)...)

	var metricsType string
	switch s.Metrics.(type) {
	case *noop.Metrics:
		metricsType = "No-Op"
	case *prometheus.Metrics:
		metricsType = "Prometheus"
	default:
		metricsType = reflect.TypeOf(s.Metrics).String()
		metricsType = strings.TrimPrefix(metricsType, "*")
	}
	lines = append(lines, subSection+"Metric type: "+metricsType)

	return lines
}
