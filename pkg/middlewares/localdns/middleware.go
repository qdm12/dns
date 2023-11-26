package localdns

import (
	"fmt"

	"github.com/miekg/dns"
)

// Middleware implements a DNS forwarder for requests
// containing a single local name question.
type Middleware struct {
	settings Settings
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

// Wrap wraps the DNS handler with the middleware.
func (m *Middleware) Wrap(next dns.Handler) dns.Handler { //nolint:ireturn
	return newHandler(m.settings.Ctx, m.settings.Resolvers,
		m.settings.Logger, next)
}
