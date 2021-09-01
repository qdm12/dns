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
	c = &counters{
		requests: helpers.NewCounter(settings.Prefix, "requests",
			"Requests received by the server"),
		questions: helpers.NewCounterVec(settings.Prefix, "questions",
			"Questions contained in requests", []string{"class", "type"}),
		rcode: helpers.NewCounterVec(settings.Prefix, "rcode",
			"Response codes", []string{"rcode"}),
		answers: helpers.NewCounterVec(settings.Prefix, "answers",
			"Answers contained in responses", []string{"class", "type"}),
		responses: helpers.NewCounter(settings.Prefix, "responses",
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
