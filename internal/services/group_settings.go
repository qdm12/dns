package services

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/services/hooks"
)

type GroupSettings struct {
	// Name is the sequence name, used for hooks and errors.
	Name string
	// Services specifies the services to start and stop in parallel.
	// Note their order does not matter.
	Services []Service
	// Hooks are hooks to call when starting and stopping
	// each service. Hooks method calls should be thread safe
	// since its methods are called in parallel goroutines.
	// It defaults to a no-op hooks implementation if left unset.
	Hooks Hooks
}

func (s *GroupSettings) SetDefaults() {
	if s.Hooks == nil {
		s.Hooks = hooks.NewNoop()
	}
}

func (s GroupSettings) Validate() (err error) {
	if len(s.Services) == 0 {
		return fmt.Errorf("%w", ErrNoService)
	}

	for i, service := range s.Services {
		if service == nil {
			return fmt.Errorf("service at index %d: %w", i, ErrServiceIsNil)
		}
	}

	errMessage := validateServicesAreUnique(s.Services)
	if errMessage != "" {
		return fmt.Errorf("%w: %s", ErrServicesNotUnique, errMessage)
	}

	return nil
}
