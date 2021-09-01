package cache

import (
	"github.com/miekg/dns"
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Interface

type Interface interface {
	Add(request, response *dns.Msg)
	Get(request *dns.Msg) (response *dns.Msg)
}
