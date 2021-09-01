// Package metrics defines the DNS metrics middleware and a
// metric interface to give to the middleware constructor.
package metrics

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/middlewares/stateful"
)

func New(settings Settings) func(dns.Handler) dns.Handler {
	settings.setDefaults()

	return func(next dns.Handler) dns.Handler {
		return &handler{
			next:    next,
			metrics: settings.Metrics,
		}
	}
}

type handler struct {
	next    dns.Handler
	metrics Interface
}

func (h *handler) ServeDNS(w dns.ResponseWriter, response *dns.Msg) {
	h.metrics.InFlightRequestsInc()
	defer h.metrics.InFlightRequestsDec()

	h.metrics.RequestsInc()

	for _, question := range response.Question {
		class := dns.Class(question.Qclass).String()
		qType := dns.Type(question.Qtype).String()
		h.metrics.QuestionsInc(class, qType)
	}

	statefulWriter := stateful.NewWriter(w)
	h.next.ServeDNS(statefulWriter, response)

	rcode := rcodeToString(statefulWriter.Response.Rcode)
	h.metrics.RcodeInc(rcode)

	for _, rr := range response.Answer {
		header := rr.Header()
		class := dns.Class(header.Class).String()
		rrType := dns.Type(header.Rrtype).String()
		h.metrics.AnswersInc(class, rrType)
	}

	h.metrics.ResponsesInc()
}

func rcodeToString(rcode int) (rcodeString string) {
	rcodeString, ok := dns.RcodeToString[rcode]
	if !ok {
		rcodeString = fmt.Sprintf("%d unknown", rcode)
	}
	return rcodeString
}
