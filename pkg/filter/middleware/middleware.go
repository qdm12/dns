package middleware

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/middlewares/stateful"
)

type Middleware struct {
	filter Filter
	cache  Cache
}

func New(filter Filter, cache Cache) *Middleware {
	return &Middleware{
		filter: filter,
		cache:  cache,
	}
}

func (m *Middleware) Wrap(next dns.Handler) dns.Handler { //nolint:ireturn
	return &handler{
		next:   next,
		filter: m.filter,
		cache:  m.cache,
	}
}

type handler struct {
	next   dns.Handler
	filter Filter
	cache  Cache
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
		h.cache.Remove(r) // remove from cache if present
		response = new(dns.Msg).SetRcode(r, dns.RcodeRefused)
		_ = w.WriteMsg(response)
		return
	}

	_ = w.WriteMsg(response)
}
