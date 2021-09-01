package metrics

import (
	"github.com/qdm12/dns/pkg/middlewares/metrics/noop"
	"github.com/qdm12/dns/pkg/middlewares/metrics/prometheus"
)

var (
	_ Interface = (*prometheus.Metrics)(nil)
	_ Interface = (*noop.Metrics)(nil)
)

type Interface interface {
	RequestsInc()
	QuestionsInc(class, qType string)
	RcodeInc(rcode string)
	AnswersInc(class, qType string)
	ResponsesInc()
	InFlightRequestsInc()
	InFlightRequestsDec()
}
