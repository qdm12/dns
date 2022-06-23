package settings

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/defaults"
	"github.com/qdm12/gotree"
)

type Prometheus struct {
	ListeningAddress string
	Subsystem        *string
}

func (p *Prometheus) setDefaults() {
	if p.ListeningAddress == "" {
		p.ListeningAddress = ":9090"
	}

	if p.Subsystem == nil {
		p.Subsystem = defaults.StringPtr(p.Subsystem, "dns")
	}
}

func (p *Prometheus) validate() (err error) {
	err = checkListeningAddress(p.ListeningAddress)
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
