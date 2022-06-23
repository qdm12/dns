package settings

import (
	"errors"
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/defaults"
	"github.com/qdm12/gotree"
)

type CacheLRU struct {
	// MaxEntries is the number of maximum entries
	// to keep in the LRU cache. It defaults to 10e4
	// if letf unset.
	MaxEntries int
}

func (c *CacheLRU) setDefaults() {
	const defaultMaxEntries = 10e4
	c.MaxEntries = defaults.Int(c.MaxEntries, defaultMaxEntries)
}

var ErrMaxEntriesNegative = errors.New("max entries must be positive")

func (c *CacheLRU) validate() error {
	if c.MaxEntries < 0 {
		return fmt.Errorf("%w: %d", ErrMaxEntriesNegative, c.MaxEntries)
	}

	return nil
}

func (c *CacheLRU) String() string {
	return c.ToLinesNode().String()
}

func (c *CacheLRU) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("LRU cache:")
	node.Appendf("Max entries: %d", c.MaxEntries)
	return node
}
