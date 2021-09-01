package noop

import (
	"strings"

	"github.com/qdm12/dns/pkg/cache/metrics"
	"github.com/qdm12/dns/pkg/cache/metrics/noop"
)

type Settings struct {
	// Metrics is the metrics interface to record the cache type.
	// It defaults to a No-Op metric implementation.
	Metrics metrics.Interface
}

func (s *Settings) setDefaults() {
	if s.Metrics == nil {
		s.Metrics = noop.New()
	}
}

func (s *Settings) String() string {
	const (
		subSection = " |--"
		indent     = "    " // used if lines already contain the subSection
	)
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *Settings) Lines(indent, subSection string) (lines []string) {
	lines = append(lines, subSection+"Cache type: "+CacheType)
	return lines
}
