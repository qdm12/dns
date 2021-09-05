// Package noop defines a No-Op metric implementation for
// the metrics middleware.
package noop

type Metrics struct{}

func New() *Metrics { return new(Metrics) }

func (m *Metrics) RequestsInc()                     {}
func (m *Metrics) QuestionsInc(class, qType string) {}
func (m *Metrics) RcodeInc(rcode string)            {}
func (m *Metrics) AnswersInc(class, qType string)   {}
func (m *Metrics) ResponsesInc()                    {}
func (m *Metrics) InFlightRequestsInc()             {}
func (m *Metrics) InFlightRequestsDec()             {}
