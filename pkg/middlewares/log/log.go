// Package log defines a middleware to handle response writing errors
// as well as log each request and its response if enabled.
package log

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/middlewares/log/format"
	formatconsole "github.com/qdm12/dns/pkg/middlewares/log/format/console"
	formatnoop "github.com/qdm12/dns/pkg/middlewares/log/format/noop"
	"github.com/qdm12/dns/pkg/middlewares/log/logger"
	lognoop "github.com/qdm12/dns/pkg/middlewares/log/logger/noop"
	"github.com/qdm12/dns/pkg/middlewares/stateful"
)

func New(settings Settings) func(dns.Handler) dns.Handler {
	settings.SetDefaults()

	var formatter format.Formatter
	switch {
	case settings.CustomFormatter != nil:
		formatter = settings.CustomFormatter
	case settings.Format == console:
		formatter = formatconsole.New()
	case settings.Format == noop:
		formatter = formatnoop.New()
	}

	var logger logger.Interface
	switch {
	case settings.CustomLogger != nil:
		logger = settings.CustomLogger
	case settings.Format == noop:
		logger = lognoop.New()
	}

	return func(next dns.Handler) dns.Handler {
		return &handler{
			formatter: formatter,
			logger:    logger,
			next:      next,
		}
	}
}

type handler struct {
	formatter format.Formatter
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
