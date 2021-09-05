// Package noop defines a No-op metric implementation for the filter.
package noop

type Metrics struct{}

func New() (metrics *Metrics) {
	return new(Metrics)
}

func (m *Metrics) SetBlockedHostnames(n int)                 {}
func (m *Metrics) SetBlockedIPs(n int)                       {}
func (m *Metrics) SetBlockedIPPrefixes(n int)                {}
func (m *Metrics) HostnamesFilteredInc(qClass, qType string) {}
func (m *Metrics) IPsFilteredInc(rrtype string)              {}
