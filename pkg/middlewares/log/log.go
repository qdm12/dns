// Package log defines a middleware to handle response writing errors
// as well as log each request and its response if enabled.
package log

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/middlewares/log/format"
	"github.com/qdm12/dns/pkg/middlewares/log/logger"
	"github.com/qdm12/dns/pkg/middlewares/stateful"
)

func New(settings Settings) func(dns.Handler) dns.Handler {
	settings.SetDefaults()

	return func(next dns.Handler) dns.Handler {
		return &handler{
			formatter: settings.Formatter,
			logger:    settings.Logger,
			next:      next,
		}
	}
}

type handler struct {
	formatter format.Interface
	logger    logger.Interface
	next      dns.Handler
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	h.logger.LogRequest(h.formatter.Request(r))

	sw := stateful.NewWriter(w)
	h.next.ServeDNS(sw, r)
	if err := sw.WriteErr; err != nil {
		errString := "cannot write DNS response: " + err.Error()
		h.logger.Error(h.formatter.Error(r.Id, errString))
	}

	h.logger.LogResponse(h.formatter.Response(sw.Response))

	h.logger.LogRequestResponse(
		h.formatter.RequestResponse(r, sw.Response))
}
