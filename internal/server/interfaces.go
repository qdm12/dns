package server

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/filter/update"
)

type Filter interface {
	FilterRequest(request *dns.Msg) (blocked bool)
	FilterResponse(response *dns.Msg) (blocked bool)
	Update(settings update.Settings)
}

type Cache interface {
	Add(request, response *dns.Msg)
	Get(request *dns.Msg) (response *dns.Msg)
	Remove(request *dns.Msg)
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
