package filter

import (
	"errors"
	"fmt"
)

type Settings struct {
	Filter Filter
}

var ErrFilterMustBeSet = errors.New("filter must be set")

func (s *Settings) Validate() (err error) {
	if s.Filter == nil {
		return fmt.Errorf("%w", ErrFilterMustBeSet)
	}

	return nil
}
