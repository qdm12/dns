// Package noop defines a No-Op metric implementation for
// the metrics middleware.
package noop

type Metrics struct{}

func New() *Metrics { return new(Metrics) }

func (m *Metrics) RequestsInc()                {}
func (m *Metrics) QuestionsInc(string, string) {}
func (m *Metrics) RcodeInc(string)             {}
func (m *Metrics) AnswersInc(string, string)   {}
func (m *Metrics) ResponsesInc()               {}
func (m *Metrics) InFlightRequestsInc()        {}
func (m *Metrics) InFlightRequestsDec()        {}
