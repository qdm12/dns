package cache

import (
	"github.com/miekg/dns"
)

type Interface interface {
	Add(request, response *dns.Msg)
	Get(request *dns.Msg) (response *dns.Msg)
	Remove(request *dns.Msg)
}
