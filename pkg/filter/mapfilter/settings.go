package mapfilter

import (
	"reflect"
	"strings"

	"github.com/qdm12/dns/pkg/filter/metrics"
	"github.com/qdm12/dns/pkg/filter/metrics/noop"
	"github.com/qdm12/dns/pkg/filter/metrics/prometheus"
	"github.com/qdm12/dns/pkg/filter/update"
	"github.com/qdm12/gotree"
)

type Settings struct {
	Update  update.Settings
	Metrics metrics.Interface
}

func (s *Settings) SetDefaults() {
	s.Update.SetDefaults()
	if s.Metrics == nil {
		s.Metrics = noop.New()
	}
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Filter settings:")
	node.AppendNode(s.Update.ToLinesNode())

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
	node.Appendf("Metrics type: %s", metricsType)

	return node
}
