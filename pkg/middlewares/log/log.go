// Package log defines a middleware to handle response writing errors
// as well as log each request and its response if enabled.
package log

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/middlewares/stateful"
)

func New(settings Settings) func(dns.Handler) dns.Handler {
	settings.SetDefaults()

	return func(next dns.Handler) dns.Handler {
		return &handler{
			logger: settings.Logger,
			next:   next,
		}
	}
}

type handler struct {
	logger Logger
	next   dns.Handler
}

type Logger interface {
	// Log logs the request and/or response.
	Log(request, response *dns.Msg)
	// Error logs errors returned by the DNS handler.
	Error(id uint16, errorString string)
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	sw := stateful.NewWriter(w)
	h.next.ServeDNS(sw, r)
	h.logger.Log(r, sw.Response)
	if err := sw.WriteErr; err != nil {
		errString := "cannot write DNS response: " + err.Error()
		h.logger.Error(r.Id, errString)
	}
}
