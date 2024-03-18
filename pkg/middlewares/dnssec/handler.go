package dnssec

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/dnssec"
)

type handler struct {
	// Injected from middleware
	logger Logger
	next   dns.Handler
}

func newHandler(logger Logger, next dns.Handler) *handler {
	return &handler{
		logger: logger,
		next:   next,
	}
}

func (h *handler) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	response, err := dnssec.Validate(request, h.next)
	if err != nil {
		h.logger.Warn(err.Error())
		response = new(dns.Msg)
		response.SetRcode(request, dns.RcodeServerFailure)
	}

	_ = w.WriteMsg(response)
}
