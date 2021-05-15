// Package log defines a middleware to handle response writing errors
// as well as log each request and its response if enabled.
package log

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/middlewares/stateful"
	"github.com/qdm12/golibs/logging"
)

func New(logger logging.Logger, settings Settings) func(dns.Handler) dns.Handler {
	return func(next dns.Handler) dns.Handler {
		return &handler{
			logger:   logger,
			next:     next,
			settings: settings,
		}
	}
}

type handler struct {
	logger   logging.Logger
	next     dns.Handler
	settings Settings
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	sw := stateful.NewWriter(w)
	h.next.ServeDNS(sw, r)
	if err := sw.WriteErr; err != nil {
		requestStr := requestString(r)
		h.logger.Error(requestStr + ": cannot write DNS response: " + err.Error())
	}

	switch {
	case h.settings.LogRequests && h.settings.LogResponses:
		requestStr := requestString(r)
		responseStr := responseString(sw.Response)
		h.logger.Info(requestStr + " => " + responseStr)
	case h.settings.LogRequests:
		requestStr := requestString(r)
		h.logger.Info(requestStr)
	case h.settings.LogResponses:
		responseStr := responseString(sw.Response)
		h.logger.Info(responseStr)
	}
}
