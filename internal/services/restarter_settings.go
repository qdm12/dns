package services

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/services/hooks"
)

type RestarterSettings struct {
	// Service is the service to restart.
	// It must be set for settings validation to succeed.
	Service Service
	// Hooks are hooks to call when the service starts,
	// stops or crashes.
	Hooks Hooks
}

func (r *RestarterSettings) SetDefaults() {
	if r.Hooks == nil {
		r.Hooks = hooks.NewNoop()
	}
}

func (r RestarterSettings) Validate() (err error) {
	if r.Service == nil {
		return fmt.Errorf("%w", ErrNoService)
	}

	return nil
}
