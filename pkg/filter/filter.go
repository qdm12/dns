package filter

import "github.com/miekg/dns"

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Filter

type Filter interface {
	FilterRequest(request *dns.Msg) (blocked bool)
	FilterResponse(response *dns.Msg) (blocked bool)
}
