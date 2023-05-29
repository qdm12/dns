package settings //nolint:dupl

import (
	"fmt"

	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
)

type Cache struct {
	// Type is the cache type, and can be
	// 'lru' or 'noop'. It defaults to 'lru' if
	// left unset.
	Type string
	LRU  CacheLRU
}

func (c *Cache) setDefaults() {
	if c.Type == "" {
		c.Type = "lru"
	}

	c.LRU.setDefaults()
}

func (c *Cache) validate() (err error) {
	err = validate.IsOneOf(c.Type, "lru", "noop")
	if err != nil {
		return fmt.Errorf("cache type: %w", err)
	}

	err = c.LRU.validate()
	if err != nil {
		return fmt.Errorf("LRU cache: %w", err)
	}

	return nil
}

func (c *Cache) String() string {
	return c.ToLinesNode().String()
}

func (c *Cache) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Cache:")
	node.Appendf("Type: %s", c.Type)
	switch c.Type {
	case "noop":
	case "lru":
		node.AppendNode(c.LRU.ToLinesNode())
	default:
		panic(fmt.Sprintf("unknown cache type: %s", c.Type))
	}
	return node
}
