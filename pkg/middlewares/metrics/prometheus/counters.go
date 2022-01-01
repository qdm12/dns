package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/pkg/metrics/prometheus"
)

type counters struct {
	requests  prometheus.Counter
	questions *prometheus.CounterVec
	rcode     *prometheus.CounterVec
	answers   *prometheus.CounterVec
	responses prometheus.Counter
}

func newCounters(settings prom.Settings) (c *counters, err error) {
	prefix := *settings.Prefix
	c = &counters{
		requests: helpers.NewCounter(prefix, "requests_received",
			"Requests received by the server"),
		questions: helpers.NewCounterVec(prefix, "questions_received",
			"Questions contained in requests received by the server", []string{"class", "type"}),
		rcode: helpers.NewCounterVec(prefix, "response_rcodes",
			"Response codes", []string{"rcode"}),
		answers: helpers.NewCounterVec(prefix, "answers_sent",
			"Answers contained in responses sent by the server", []string{"class", "type"}),
		responses: helpers.NewCounter(prefix, "responses_sent",
			"Responses sent out by the server"),
	}

	err = helpers.Register(settings.Registry, c.requests, c.questions,
		c.rcode, c.answers, c.responses)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *counters) RequestsInc() {
	c.requests.Inc()
}

func (c *counters) QuestionsInc(class, qType string) {
	c.questions.WithLabelValues(class, qType).Inc()
}

func (c *counters) RcodeInc(rcode string) {
	c.rcode.WithLabelValues(rcode).Inc()
}

func (c *counters) AnswersInc(class, qType string) {
	c.answers.WithLabelValues(class, qType).Inc()
}

func (c *counters) ResponsesInc() {
	c.responses.Inc()
}
