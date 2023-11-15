package cache

import (
	"errors"
	"fmt"
)

type Settings struct {
	Cache Cache
}

var (
	ErrCacheMustBeSet = errors.New("cache must be set")
)

func (s *Settings) Validate() (err error) {
	if s.Cache == nil {
		return fmt.Errorf("%w", ErrCacheMustBeSet)
	}

	return nil
}
