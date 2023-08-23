package middleware

import "github.com/miekg/dns"

type Filter interface {
	FilterRequest(request *dns.Msg) bool
	FilterResponse(response *dns.Msg) bool
}
