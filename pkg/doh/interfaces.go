package doh

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/provider"
)

type Middleware interface {
	String() string
	Wrap(next dns.Handler) dns.Handler
	Stop() (err error)
}

type Metrics interface {
	DoHDialInc(url string)
	DoTDialInc(provider, address, outcome string)
	DNSDialInc(address, outcome string)
}

type Logger interface {
	Debug(s string)
	Info(s string)
	Warner
	Error(s string)
}

type Warner interface {
	Warn(s string)
}

type Picker interface {
	DoHServer(servers []provider.DoHServer) provider.DoHServer
}
