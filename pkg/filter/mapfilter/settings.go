package mapfilter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/qdm12/dns/v2/pkg/filter/metrics"
	"github.com/qdm12/dns/v2/pkg/filter/metrics/noop"
	"github.com/qdm12/dns/v2/pkg/filter/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/filter/update"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// Update contains the filter update settings.
	Update update.Settings
	// Metrics is the metric interface and defaults
	// to a no-op implementation if left unset.
	Metrics metrics.Interface
}

func (s *Settings) SetDefaults() {
	s.Update.SetDefaults()

	if s.Metrics == nil {
		s.Metrics = noop.New()
	}
}

func (s Settings) Validate() (err error) {
	err = s.Update.Validate()
	if err != nil {
		return fmt.Errorf("failed validating update settings: %w", err)
	}

	return nil
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
