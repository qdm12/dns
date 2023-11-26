package localdns

import (
	"fmt"

	"github.com/miekg/dns"
)

// Middleware implements a DNS forwarder for requests
// containing a single local name question.
type Middleware struct {
	settings Settings
	handlers []*handler
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
	return "local DNS"
}

// Wrap wraps the DNS handler with the middleware.
func (m *Middleware) Wrap(next dns.Handler) dns.Handler { //nolint:ireturn
	handler := newHandler(m.settings.Resolvers,
		m.settings.Logger, next)
	m.handlers = append(m.handlers, handler)
	return handler
}

// Stop stops the middleware, and all wrapping DNS handlers
// created by the middleware will cease to handle requests.
// The function returns once all handlers are done with their
// previously ongoing ServeDNS calls.
func (m *Middleware) Stop() (err error) {
	for _, handler := range m.handlers {
		handler.stop()
	}
	return nil
}
