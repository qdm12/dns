package setup

import (
	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/pkg/filter/update"
)

type Logger interface {
	Debug(s string)
	Info(s string)
	Warn(s string)
	Error(s string)
}

type Filter interface {
	FilterRequest(request *dns.Msg) (blocked bool)
	FilterResponse(response *dns.Msg) (blocked bool)
	Update(settings update.Settings)
}

type PrometheusRegisterer prometheus.Registerer

type Metrics interface {
	DoTMetrics
	DoHMetrics
	MiddlewareMetrics
}

type DoTMetrics interface {
	DoTDialInc(provider, address, outcome string)
	DNSDialInc(address, outcome string)
	MiddlewareMetrics
}

type DoHMetrics interface {
	DoHDialInc(url string)
	DoTDialInc(provider, address, outcome string)
	DNSDialInc(address, outcome string)
	MiddlewareMetrics
}

type MiddlewareMetrics interface {
	RequestsInc()
	QuestionsInc(class, qType string)
	RcodeInc(rcode string)
	AnswersInc(class, qType string)
	ResponsesInc()
	InFlightRequestsInc()
	InFlightRequestsDec()
}
