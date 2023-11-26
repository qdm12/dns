package cache

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/stateful"
)

type Middleware struct {
	cache Cache
}

func New(settings Settings) (middleware *Middleware, err error) {
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("settings validation: %w", err)
	}

	return &Middleware{
		cache: settings.Cache,
	}, nil
}

func (m *Middleware) String() string {
	return "cache"
}

func (m *Middleware) Wrap(next dns.Handler) dns.Handler { //nolint:ireturn
	return &handler{
		next:  next,
		cache: m.cache,
	}
}

func (m *Middleware) Stop() (err error) {
	return nil
}

type handler struct {
	next  dns.Handler
	cache Cache
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	response := h.cache.Get(r)
	if response != nil {
		response.SetReply(r)
		_ = w.WriteMsg(response)
		return
	}

	statefulWriter := stateful.NewWriter()
	h.next.ServeDNS(statefulWriter, r)
	response = statefulWriter.Response

	h.cache.Add(r, response)

	_ = w.WriteMsg(response)
}
