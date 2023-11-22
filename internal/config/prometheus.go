package config

import (
	"fmt"
	"os"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
)

type Prometheus struct {
	ListeningAddress string
	Subsystem        *string
}

func (p *Prometheus) SetDefaults() {
	p.ListeningAddress = gosettings.DefaultComparable(p.ListeningAddress, ":9090")
	p.Subsystem = gosettings.DefaultPointer(p.Subsystem, "dns")
}

func (p *Prometheus) Validate() (err error) {
	err = validate.ListeningAddress(p.ListeningAddress, os.Getuid())
	if err != nil {
		return fmt.Errorf("listening address: %w", err)
	}

	return nil
}

func (p *Prometheus) String() string {
	return p.ToLinesNode().String()
}

func (p *Prometheus) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Prometheus:")
	node.Appendf("Listening address: %s", p.ListeningAddress)
	if *p.Subsystem != "" {
		node.Appendf("Subsystem: %s", *p.Subsystem)
	}
	return node
}

func (p *Prometheus) read(reader *reader.Reader) {
	p.ListeningAddress = reader.String("METRICS_PROMETHEUS_ADDRESS")
	p.Subsystem = reader.Get("METRICS_PROMETHEUS_SUBSYSTEM")
}
