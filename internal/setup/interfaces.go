package setup

import (
	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/pkg/filter/update"
)

type Logger interface {
	Debug(s string)
	Info(s string)
	Warn(s string)
	Error(s string)
}

type Filter interface {
	FilterRequest(request *dns.Msg) (blocked bool)
	FilterResponse(response *dns.Msg) (blocked bool)
	Update(settings update.Settings)
}

type FilterMetrics interface {
	SetBlockedHostnames(n int)
	SetBlockedIPs(n int)
	SetBlockedIPPrefixes(n int)
	HostnamesFilteredInc(qClass, qType string)
	IPsFilteredInc(rrtype string)
}

type Middleware interface {
	Wrap(next dns.Handler) dns.Handler
}

type PrometheusRegisterer prometheus.Registerer

type DoTMetrics interface {
	DoTDialInc(provider, address, outcome string)
	DNSDialInc(address, outcome string)
}

type DoHMetrics interface {
	DoHDialInc(url string)
	DoTDialInc(provider, address, outcome string)
	DNSDialInc(address, outcome string)
}

type CacheMetrics interface { //nolint:interfacebloat
	SetCacheType(cacheType string)
	CacheInsertInc()
	CacheRemoveInc()
	CacheMoveInc()
	CacheGetEmptyInc()
	CacheInsertEmptyInc()
	CacheRemoveEmptyInc()
	CacheHitInc()
	CacheMissInc()
	CacheExpiredInc()
	CacheMaxEntriesSet(maxEntries uint)
}
