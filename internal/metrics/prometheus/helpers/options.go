package helpers

import (
	"github.com/prometheus/client_golang/prometheus"
)

func newOpts(prefix, name, help string) (opts prometheus.Opts) {
	opts.Subsystem = prefix
	opts.Name = name
	opts.Help = help
	return opts
}
