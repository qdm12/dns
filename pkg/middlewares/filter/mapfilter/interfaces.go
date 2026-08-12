package mapfilter

type LocalChecker interface {
	IsFQDNLocal(fqdn string) bool
}

type Metrics interface {
	SetBlockedHostnames(n int)
	SetBlockedIPs(n int)
	SetBlockedIPPrefixes(n int)
	SetFqdnExemptFromRebindingProtection(n int)
	HostnamesFilteredInc(qClass, qType string)
	IPsFilteredInc(rrtype string)
}

type Logger interface {
	Log(s string)
}
