package dot

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/filter/update"
)

type middlewareMetrics interface { //nolint:unused
	RequestsInc()
	QuestionsInc(class, qType string)
	RcodeInc(rcode string)
	AnswersInc(class, qType string)
	ResponsesInc()
	InFlightRequestsInc()
	InFlightRequestsDec()
}

type filter interface { //nolint:unused
	FilterRequest(request *dns.Msg) (blocked bool)
	FilterResponse(response *dns.Msg) (blocked bool)
	Update(settings update.Settings)
}

type cache interface { //nolint:unused
	Add(request, response *dns.Msg)
	Get(request *dns.Msg) (response *dns.Msg)
	Remove(request *dns.Msg)
}
