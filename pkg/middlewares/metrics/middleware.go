// Package metrics defines the DNS metrics middleware and a
// metric interface to give to the middleware constructor.
package metrics

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/stateful"
)

type Middleware struct {
	metrics Metrics
}

func New(settings Settings) (middleware *Middleware, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("settings validation: %w", err)
	}

	return &Middleware{
		metrics: settings.Metrics,
	}, nil
}

func (m *Middleware) String() string {
	return "metrics"
}

// Wrap wraps the DNS handler with the middleware.
func (m *Middleware) Wrap(next dns.Handler) dns.Handler { //nolint:ireturn
	return &handler{
		next:    next,
		metrics: m.metrics,
	}
}

func (m *Middleware) Stop() (err error) {
	return nil
}

type handler struct {
	next    dns.Handler
	metrics Metrics
}

func (h *handler) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	h.metrics.InFlightRequestsInc()
	defer h.metrics.InFlightRequestsDec()

	h.metrics.RequestsInc()

	for _, question := range request.Question {
		class := dns.Class(question.Qclass).String()
		qType := dns.Type(question.Qtype).String()
		h.metrics.QuestionsInc(class, qType)
	}

	statefulWriter := stateful.NewWriter()
	h.next.ServeDNS(statefulWriter, request)
	response := statefulWriter.Response

	rcode := rcodeToString(response.Rcode)
	h.metrics.RcodeInc(rcode)

	for _, rr := range response.Answer {
		header := rr.Header()
		class := dns.Class(header.Class).String()
		rrType := dns.Type(header.Rrtype).String()
		h.metrics.AnswersInc(class, rrType)
	}

	h.metrics.ResponsesInc()

	_ = w.WriteMsg(statefulWriter.Response)
}

func rcodeToString(rcode int) (rcodeString string) {
	rcodeString, ok := dns.RcodeToString[rcode]
	if !ok {
		rcodeString = fmt.Sprintf("%d unknown", rcode)
	}
	return rcodeString
}
