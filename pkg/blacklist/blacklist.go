package blacklist

import "github.com/miekg/dns"

type BlackLister interface {
	FilterRequest(request *dns.Msg) (blocked bool)
	FilterResponse(response *dns.Msg) (blocked bool)
}
