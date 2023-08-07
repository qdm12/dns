package noop

import (
	"github.com/miekg/dns"
)

type NoOp struct {
	metrics Metrics
}

func New(settings Settings) *NoOp {
	settings.SetDefaults()
	settings.Metrics.SetCacheType(CacheType)
	return &NoOp{
		metrics: settings.Metrics,
	}
}

func (n *NoOp) Add(*dns.Msg, *dns.Msg)           {}
func (n *NoOp) Get(*dns.Msg) (response *dns.Msg) { return nil }
func (n *NoOp) Remove(*dns.Msg)                  {}
