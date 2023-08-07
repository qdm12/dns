package prometheus

type DialMetrics interface {
	DoTDialInc(provider, address, outcome string)
	DNSDialInc(address, outcome string)
}
