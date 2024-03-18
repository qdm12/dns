package dnssec

import (
	"fmt"
	"sync/atomic"

	"github.com/miekg/dns"
)

// Middleware implements a DNSSEC validator.
type Middleware struct {
	settings Settings
	wrapping atomic.Bool
}

func New(settings Settings) (middleware *Middleware, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	return &Middleware{
		settings: settings,
	}, nil
}

func (m *Middleware) String() string {
	return "DNSSEC validator"
}

// Wrap wraps the DNS handler with the middleware.
func (m *Middleware) Wrap(next dns.Handler) dns.Handler { //nolint:ireturn
	previousWrapping := m.wrapping.Swap(true)
	if previousWrapping {
		panic("DNSSEC middleware cannot wrap more than once")
	}

	handler := newHandler(m.settings.Logger, next)
	return handler
}

// Stop is a no-op for the DNSSEC middleware.
func (m *Middleware) Stop() (err error) {
	return nil
}
