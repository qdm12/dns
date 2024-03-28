package substituter

import (
	"fmt"

	"github.com/miekg/dns"
)

type Middleware struct {
	mapping map[questionKey][]dns.RR
}

func New(settings Settings) (middleware *Middleware, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	mapping := make(map[questionKey][]dns.RR, len(settings.Substitutions))
	for _, substitution := range settings.Substitutions {
		substitution.setDefaults()
		question := substitution.toQuestion()
		key := makeKey(question)
		mapping[key] = substitution.toRRs()
	}

	return &Middleware{
		mapping: mapping,
	}, nil
}

func (m *Middleware) String() string { return "substituter" }

// Wrap wraps the DNS handler with the middleware.
func (m *Middleware) Wrap(next dns.Handler) dns.Handler { //nolint:ireturn
	if len(m.mapping) == 0 {
		return next
	}
	return &handler{
		mapping: m.mapping,
		next:    next,
	}
}

func (m *Middleware) Stop() (err error) {
	return nil
}

type handler struct {
	mapping map[questionKey][]dns.RR
	next    dns.Handler
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	for _, question := range r.Question {
		key := makeKey(question)
		rrs, found := h.mapping[key]
		if !found {
			continue
		}
		response := &dns.Msg{
			Answer: rrs,
		}
		response.SetReply(r)
		_ = w.WriteMsg(response)
		return
	}

	h.next.ServeDNS(w, r)
}

func makeKey(question dns.Question) (key questionKey) {
	return questionKey{
		Name:   question.Name,
		Qtype:  question.Qtype,
		Qclass: question.Qclass,
	}
}

type questionKey struct {
	Name   string
	Qtype  uint16
	Qclass uint16
}
