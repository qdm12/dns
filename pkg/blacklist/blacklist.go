package blacklist

import "github.com/miekg/dns"

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . BlackLister

type BlackLister interface {
	FilterRequest(request *dns.Msg) (blocked bool)
	FilterResponse(response *dns.Msg) (blocked bool)
}
