package cache

import "github.com/miekg/dns"

type Cache interface {
	Get(request *dns.Msg) *dns.Msg
	Add(request, response *dns.Msg)
}
