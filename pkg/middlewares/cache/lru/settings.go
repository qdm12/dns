package lru

import (
	"errors"
	"fmt"

	"github.com/qdm12/dns/v2/pkg/middlewares/cache/metrics/noop"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// MaxEntries is the maximum number of request<->response pairs
	// to be stored in the cache. It defaults to 10e4 if left unset.
	// Note its type is int insted of uint* since its maximum value
	// is math.MaxInt (it's used as an int length).
	MaxEntries int
	// Metrics is the metrics interface to record metric information
	// for the cache. It defaults to a No-Op metric implementation.
	Metrics Metrics
}

func (s *Settings) SetDefaults() {
	const defaultMaxEntries = 10e4
	s.MaxEntries = gosettings.DefaultComparable(s.MaxEntries, defaultMaxEntries)
	s.Metrics = gosettings.DefaultComparable[Metrics](s.Metrics, noop.New())
}

var (
	ErrMaxEntriesNegative = errors.New("max entries is negative")
	ErrMaxEntriesZero     = errors.New("max entries is zero")
)

func (s Settings) Validate() (err error) {
	switch {
	case s.MaxEntries < 0:
		return fmt.Errorf("%w: %d", ErrMaxEntriesNegative, s.MaxEntries)
	case s.MaxEntries == 0:
		return fmt.Errorf("%w", ErrMaxEntriesZero)
	}
	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("LRU cache settings:")
	node.Appendf("Max entries: %d", s.MaxEntries)
	return node
}
