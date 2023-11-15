// Package log defines a middleware to handle response writing errors
// as well as log each request and its response if enabled.
package log

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/stateful"
)

type Middleware struct {
	logger Logger
}

func New(settings Settings) (middleware *Middleware, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("settings validation: %w", err)
	}

	return &Middleware{
		logger: settings.Logger,
	}, nil
}

// Wrap wraps the DNS handler with the middleware.
func (m *Middleware) Wrap(next dns.Handler) dns.Handler { //nolint:ireturn
	return &handler{
		logger: m.logger,
		next:   next,
	}
}

type handler struct {
	logger Logger
	next   dns.Handler
}

type Logger interface {
	// Log logs the request and/or response.
	Log(remoteAddr net.Addr, request, response *dns.Msg)
	// Error logs errors returned by the DNS handler.
	Error(id uint16, errorString string)
}

// ServeDNS implements the dns.Handler interface.
// Note the response writer passed as argument should be an actual
// IO writer, not a buffered writer, so it can return an actual error.
func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	sw := stateful.NewWriter()
	h.next.ServeDNS(sw, r)
	h.logger.Log(w.RemoteAddr(), r, sw.Response)
	err := w.WriteMsg(sw.Response)
	if err != nil {
		errString := "cannot write DNS response: " + err.Error()
		h.logger.Error(r.Id, errString)
	}
}
