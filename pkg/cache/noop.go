package cache

import (
	"github.com/miekg/dns"
)

type noop struct{}

func newNoop() *noop {
	return &noop{}
}

func (n *noop) Add(request, response *dns.Msg)           {}
func (n *noop) Get(request *dns.Msg) (response *dns.Msg) { return nil }
