package mapfilter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/qdm12/dns/v2/pkg/middlewares/filter/metrics/noop"
	"github.com/qdm12/dns/v2/pkg/middlewares/filter/metrics/prometheus"
	"github.com/qdm12/dns/v2/pkg/middlewares/filter/update"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// Update contains the filter update settings.
	Update update.Settings
	// Metrics is the metric interface and defaults
	// to a no-op implementation if left unset.
	Metrics Metrics
	// Logger logs all filtered requests and responses and why
	// they were filtered. It defaults to a no-op logger.
	Logger Logger
}

func (s *Settings) SetDefaults() {
	s.Update.SetDefaults()
	s.Metrics = gosettings.DefaultComparable[Metrics](s.Metrics, noop.New())
	s.Logger = gosettings.DefaultComparable[Logger](s.Logger, &noopLogger{})
}

type noopLogger struct{}

func (l noopLogger) Log(_ string) {}

func (s Settings) Validate() (err error) {
	err = s.Update.Validate()
	if err != nil {
		return fmt.Errorf("update settings: %w", err)
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

	var loggerType string
	switch s.Logger.(type) {
	case *noopLogger:
		loggerType = "No-Op"
	default:
		loggerType = reflect.TypeOf(s.Logger).String()
		loggerType = strings.TrimPrefix(loggerType, "*")
	}
	node.Appendf("Logger type: %s", loggerType)

	return node
}
