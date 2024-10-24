// Package noop defines a No-Op metric implementation for DoT.
package noop

type Metrics struct{}

func New() *Metrics {
	return &Metrics{}
}

func (m *Metrics) PlainDialInc(string, string) {}
