// Package prometheus defines shared elements for Prometheus.
package prometheus

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/gosettings"
)

type Settings struct {
	// Prefix, aka Subsystem, is the prefix string in front
	// of all metric names.
	Prefix string
	// Registry is the Prometheus registry to use for the metrics.
	// It defaults to prometheus.DefaultRegisterer if left unset.
	Registry Registry
}

func (s *Settings) SetDefaults() {
	s.Registry = gosettings.DefaultInterface(s.Registry, prometheus.DefaultRegisterer)
}

var (
	ErrPrefixContainsSpace = errors.New("prefix contains one or more spaces")
)

func (s Settings) Validate() (err error) {
	if strings.Contains(s.Prefix, " ") {
		return fmt.Errorf("%w: %s", ErrPrefixContainsSpace, s.Prefix)
	}

	return nil
}
