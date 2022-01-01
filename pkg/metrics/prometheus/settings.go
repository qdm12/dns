// Package prometheus defines shared elements for Prometheus.
package prometheus

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type Settings struct {
	// Prefix, aka Subsystem, is the prefix string in front
	// of all metric names.
	// It cannot be nil in the internal state.
	Prefix *string
	// Registry is the Prometheus registerer to use for the metrics.
	// It defaults to prometheus.DefaultRegisterer if unset.
	Registry prometheus.Registerer
}

func (s *Settings) SetDefaults() {
	if s.Prefix == nil {
		prefix := ""
		s.Prefix = &prefix
	}

	if s.Registry == nil {
		s.Registry = prometheus.DefaultRegisterer
	}
}

var (
	ErrPrefixContainsSpace = errors.New("prefix contains one or more spaces")
)

func (s Settings) Validate() (err error) {
	if strings.Contains(*s.Prefix, " ") {
		return fmt.Errorf("%w: %s", ErrPrefixContainsSpace, *s.Prefix)
	}

	return nil
}
