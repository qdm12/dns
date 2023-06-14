package settings

import (
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type CacheLRU struct {
	// MaxEntries is the number of maximum entries
	// to keep in the LRU cache. It defaults to 10e4
	// if letf unset.
	MaxEntries uint
}

func (c *CacheLRU) setDefaults() {
	const defaultMaxEntries = 10e4
	c.MaxEntries = gosettings.DefaultNumber(c.MaxEntries, defaultMaxEntries)
}

func (c *CacheLRU) validate() error {
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
