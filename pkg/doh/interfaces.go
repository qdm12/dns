package doh

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/provider"
)

type Middleware interface {
	Wrap(next dns.Handler) dns.Handler
}

type Picker interface {
	DoHServer(servers []provider.DoHServer) provider.DoHServer
}
