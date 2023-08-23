package middleware

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/stateful"
)

type Middleware struct {
	filter Filter
}

func New(filter Filter) *Middleware {
	return &Middleware{
		filter: filter,
	}
}

func (m *Middleware) Wrap(next dns.Handler) dns.Handler { //nolint:ireturn
	return &handler{
		next:   next,
		filter: m.filter,
	}
}

type handler struct {
	next   dns.Handler
	filter Filter
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if h.filter.FilterRequest(r) {
		response := new(dns.Msg).SetRcode(r, dns.RcodeRefused)
		_ = w.WriteMsg(response)
		return
	}

	statefulWriter := stateful.NewWriter()
	// Note the next.ServeDNS call might retrieve a response
	// from the cache.
	h.next.ServeDNS(statefulWriter, r)
	response := statefulWriter.Response

	if h.filter.FilterResponse(response) {
		response = new(dns.Msg).SetRcode(r, dns.RcodeRefused)
		_ = w.WriteMsg(response)
		return
	}

	_ = w.WriteMsg(response)
}
