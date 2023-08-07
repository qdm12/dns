// Package noop defines a No-op metric implementation for DoH.
package noop

type Metrics struct{}

func New() (metrics *Metrics) {
	return &Metrics{}
}

func (m *Metrics) DNSDialInc(_, _ string)    {}
func (m *Metrics) DoTDialInc(_, _, _ string) {}
func (m *Metrics) DoHDialInc(string)         {}
