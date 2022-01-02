package noop

import (
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/cache/metrics"
)

type NoOp struct {
	metrics metrics.Interface
}

func New(settings Settings) *NoOp {
	settings.SetDefaults()
	settings.Metrics.SetCacheType(CacheType)
	return &NoOp{
		metrics: settings.Metrics,
	}
}

func (n *NoOp) Add(request, response *dns.Msg)           {}
func (n *NoOp) Get(request *dns.Msg) (response *dns.Msg) { return nil }
