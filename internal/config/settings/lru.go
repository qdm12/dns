package settings

import (
	"errors"
	"fmt"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type CacheLRU struct {
	// MaxEntries is the number of maximum entries
	// to keep in the LRU cache. It defaults to 10e4
	// if letf unset.
	// Note its type is int instead of uint* since it is
	// used to compare against an int length so its maximum
	// is math.MaxInt which depends on the platform.
	MaxEntries int
}

func (c *CacheLRU) setDefaults() {
	const defaultMaxEntries = 10e4
	c.MaxEntries = gosettings.DefaultComparable(c.MaxEntries, defaultMaxEntries)
}

var (
	ErrCacheLRUMaxEntriesNegative = errors.New("max entries is negative")
	ErrCacheLRUMaxEntriesZero     = errors.New("max entries is zero")
)

func (c *CacheLRU) validate() error {
	switch {
	case c.MaxEntries < 0:
		return fmt.Errorf("%w: %d", ErrCacheLRUMaxEntriesNegative, c.MaxEntries)
	case c.MaxEntries == 0:
		return fmt.Errorf("%w", ErrCacheLRUMaxEntriesZero)
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
