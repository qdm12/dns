package mapfilter

import (
	"github.com/qdm12/dns/pkg/filter/metrics"
	"inet.af/netaddr"
)

type Filter struct {
	fqdnHostnames map[string]struct{}
	ips           map[netaddr.IP]struct{}
	ipPrefixes    []netaddr.IPPrefix
	metrics       metrics.Interface
}

func New(settings Settings) *Filter {
	settings.setDefaults()
	metrics := settings.Metrics

	fqdnHostnamesSet := make(map[string]struct{}, len(settings.FqdnHostnames))
	for _, fqdnHostname := range settings.FqdnHostnames {
		fqdnHostnamesSet[fqdnHostname] = struct{}{}
	}
	metrics.SetBlockedHostnames(len(fqdnHostnamesSet))

	ipsSet := make(map[netaddr.IP]struct{}, len(settings.IPs))
	for _, ip := range settings.IPs {
		ipsSet[ip] = struct{}{}
	}
	metrics.SetBlockedIPs(len(ipsSet))

	metrics.SetBlockedIPPrefixes(len(settings.IPPrefixes))

	return &Filter{
		fqdnHostnames: fqdnHostnamesSet,
		ips:           ipsSet,
		ipPrefixes:    settings.IPPrefixes,
		metrics:       metrics,
	}
}
