package mapfilter

type Metrics interface {
	SetBlockedHostnames(n int)
	SetBlockedIPs(n int)
	SetBlockedIPPrefixes(n int)
	HostnamesFilteredInc(qClass, qType string)
	IPsFilteredInc(rrtype string)
}
