// Package noop defines a No-op metric implementation for the filter.
package noop

type Metrics struct{}

func New() (metrics *Metrics) {
	return new(Metrics)
}

func (m *Metrics) SetBlockedHostnames(int)             {}
func (m *Metrics) SetBlockedIPs(int)                   {}
func (m *Metrics) SetBlockedIPPrefixes(int)            {}
func (m *Metrics) HostnamesFilteredInc(string, string) {}
func (m *Metrics) IPsFilteredInc(string)               {}
