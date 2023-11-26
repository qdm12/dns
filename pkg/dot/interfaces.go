package dot

import (
	"github.com/miekg/dns"
)

type Middleware interface {
	String() string
	Wrap(next dns.Handler) dns.Handler
	Stop() (err error)
}

type Metrics interface {
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
