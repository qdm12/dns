package lru

import (
	"strconv"
	"strings"

	"github.com/qdm12/dns/pkg/cache/metrics"
	"github.com/qdm12/dns/pkg/cache/metrics/noop"
)

type Settings struct {
	// MaxEntries is the maximum number of request<->response pairs
	// to be stored in the cache. It defaults to 10e4 if left unset.
	MaxEntries int
	// Metrics is the metrics interface to record metric information
	// for the cache. It defaults to a No-Op metric implementation.
	Metrics metrics.Interface
}

func (s *Settings) setDefaults() {
	if s.MaxEntries == 0 {
		s.MaxEntries = 10e4
	}

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
	lines = append(lines, subSection+"Max entries: "+strconv.Itoa(s.MaxEntries))
	return lines
}
