package lru

import (
	"github.com/qdm12/dns/pkg/cache/metrics"
	"github.com/qdm12/dns/pkg/cache/metrics/noop"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// MaxEntries is the maximum number of request<->response pairs
	// to be stored in the cache. It defaults to 10e4 if left unset.
	MaxEntries int
	// Metrics is the metrics interface to record metric information
	// for the cache. It defaults to a No-Op metric implementation.
	Metrics metrics.Interface
}

func (s *Settings) SetDefaults() {
	if s.MaxEntries == 0 {
		s.MaxEntries = 10e4
	}

	if s.Metrics == nil {
		s.Metrics = noop.New()
	}
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("LRU cache settings:")
	node.Appendf("Max entries: %d", s.MaxEntries)
	return node
}
