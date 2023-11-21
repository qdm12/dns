package settings //nolint:dupl

import (
	"fmt"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
)

type Metrics struct {
	Type       string
	Prometheus Prometheus
}

func (m *Metrics) setDefaults() {
	m.Type = gosettings.DefaultComparable(m.Type, "prometheus")
	m.Prometheus.SetDefaults()
}

func (m *Metrics) validate() (err error) {
	err = validate.IsOneOf(m.Type, "prometheus", "noop")
	if err != nil {
		return fmt.Errorf("metrics type: %w", err)
	}

	err = m.Prometheus.Validate()
	if err != nil {
		return fmt.Errorf("prometheus metrics: %w", err)
	}

	return nil
}

func (m *Metrics) String() string {
	return m.ToLinesNode().String()
}

func (m *Metrics) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Metrics:")

	node.Appendf("Type: %s", m.Type)

	switch m.Type {
	case "noop":
	case "prometheus":
		node.AppendNode(m.Prometheus.ToLinesNode())
	default:
		panic(fmt.Sprintf("unknown metrics type: %s", m.Type))
	}

	return node
}
