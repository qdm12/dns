package services

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/services/hooks"
)

type SequenceSettings struct {
	// Name is the sequence name, used for hooks and errors.
	Name string
	// ServicesStart specifies an order of services
	// to start and must be set.
	ServicesStart []Service
	// ServicesStart specifies an order of services
	// to stop and must be set.
	ServicesStop []Service
	// Hooks are hooks to call when starting and stopping
	// each service.
	Hooks Hooks
}

func (s *SequenceSettings) SetDefaults() {
	if s.Hooks == nil {
		s.Hooks = hooks.NewNoop()
	}
}

func (s SequenceSettings) Validate() (err error) {
	switch {
	case len(s.ServicesStart) == 0:
		return fmt.Errorf("%w", ErrNoServiceStart)
	case len(s.ServicesStop) == 0:
		return fmt.Errorf("%w", ErrNoServiceStop)
	case len(s.ServicesStart) != len(s.ServicesStop):
		return fmt.Errorf("%w: %d services to start (%s) and %d services to stop (%s)",
			ErrServicesStartStopMismatch, len(s.ServicesStart), andServiceStrings(s.ServicesStart),
			len(s.ServicesStop), andServiceStrings(s.ServicesStop))
	}

	for i, service := range s.ServicesStart {
		if service == nil {
			return fmt.Errorf("service to start at index %d: %w", i, ErrServiceIsNil)
		}
	}

	for i, service := range s.ServicesStop {
		if service == nil {
			return fmt.Errorf("service to stop at index %d: %w", i, ErrServiceIsNil)
		}
	}

	errMessage := validateServicesStartStopMatch(s.ServicesStart, s.ServicesStop)
	if errMessage != "" {
		return fmt.Errorf("%w: %s", ErrServicesStartStopMismatch, errMessage)
	}

	errMessage = validateServicesAreUnique(s.ServicesStart)
	if errMessage != "" {
		return fmt.Errorf("%w: %s", ErrServicesNotUnique, errMessage)
	}

	return nil
}

func validateServicesStartStopMatch(servicesStart, servicesStop []Service) (errMessage string) {
	match := true
	for _, serviceStart := range servicesStart {
		serviceStartFound := false
		for _, serviceStop := range servicesStop {
			if serviceStart == serviceStop {
				serviceStartFound = true
				break
			}
		}
		if !serviceStartFound {
			match = false
			break
		}
	}

	if match {
		return ""
	}

	if len(servicesStart) == 1 {
		return fmt.Sprintf("service to start %s is not the service to stop %s",
			servicesStart[0], servicesStop[0])
	}

	return fmt.Sprintf("services to start %s are not the services to stop %s",
		andServiceStrings(servicesStart), andServiceStrings(servicesStop))
}
