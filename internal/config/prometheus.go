package config

import (
	"fmt"

	"github.com/qdm12/golibs/params"
	"github.com/qdm12/gotree"
)

type Prometheus struct {
	// Server listening address for prometheus server.
	Address   string
	Subsystem string
}

func getPrometheusSettings(reader *Reader) (settings Prometheus,
	err error) {
	var warning string
	settings.Address, warning, err = reader.env.ListeningAddress(
		"METRICS_PROMETHEUS_ADDRESS", params.Default(":9090"))
	if warning != "" {
		reader.logger.Warn("METRICS_PROMETHEUS_ADDRESS: " + warning)
	}
	if err != nil {
		return settings, fmt.Errorf("environment variable METRICS_PROMETHEUS_ADDRESS: %w", err)
	}

	settings.Subsystem, err = reader.env.Get("METRICS_PROMETHEUS_SUBSYSTEM", params.Default("dns"))
	if err != nil {
		return settings, fmt.Errorf("environment variable METRICS_PROMETHEUS_SUBSYSTEM: %w", err)
	}

	return settings, nil
}

func (p *Prometheus) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Prometheus settings:")
	node.Appendf("Listening address: %s", p.Address)
	node.Appendf("Subsystem: %s", p.Subsystem)
	return node
}
