package dot

import (
	"context"
	"crypto/tls"
	"net"
	"sync/atomic"

	"github.com/qdm12/dns/pkg/provider"
)

// NewResolver creates a DNS over TLS resolver.
func NewResolver(options ...Option) *net.Resolver {
	settings := defaultSettings()
	for _, option := range options {
		option(&settings)
	}

	dialer := &net.Dialer{
		Timeout: settings.timeout,
	}

	picker := newPicker()

	return &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			DoTServer := picker.DoTServer(settings.dotServers)
			ip := picker.DoTIP(DoTServer, settings.ipv6)
			tlsAddr := net.JoinHostPort(ip.String(), "853")

			conn, err := dialer.DialContext(ctx, "tcp", tlsAddr)
			if err != nil {
				if len(settings.dnsServers) > 0 {
					// fallback on plain DNS if DoT does not work
					dnsServer := picker.DNSServer(settings.dnsServers)
					ip := picker.DNSIP(dnsServer, settings.ipv6)
					plainAddr := net.JoinHostPort(ip.String(), "53")
					return dialer.DialContext(ctx, "udp", plainAddr)
				}
				return nil, err
			}

			tlsConf := &tls.Config{
				MinVersion: tls.VersionTLS12,
				ServerName: DoTServer.Name,
			}
			// TODO handshake? See tls.DialWithDialer
			return tls.Client(conn, tlsConf), nil
		},
	}
}

type picker struct {
	counterDoTServer *int64
	counterDoTIPv4   *int64
	counterDoTIPv6   *int64
	counterDNSServer *int64
	counterDNSIPv4   *int64
	counterDNSIPv6   *int64
}

func newPicker() *picker {
	return &picker{
		counterDoTServer: new(int64),
		counterDoTIPv4:   new(int64),
		counterDoTIPv6:   new(int64),
		counterDNSServer: new(int64),
		counterDNSIPv4:   new(int64),
		counterDNSIPv6:   new(int64),
	}
}

func (p *picker) randIndex(counter *int64, max int) int {
	index := int(atomic.AddInt64(counter, 1))
	return index % max
}

func (p *picker) DoTServer(servers []provider.DoTServer) provider.DoTServer {
	index := 0
	if nServers := len(servers); nServers > 1 {
		index = p.randIndex(p.counterDoTServer, nServers)
	}
	return servers[index]
}

func (p *picker) IP(counter *int64, ips []net.IP) net.IP {
	index := 0
	if nIPs := len(ips); nIPs > 1 {
		index = p.randIndex(counter, nIPs)
	}
	return ips[index]
}

func (p *picker) DoTIP(server provider.DoTServer, ipv6 bool) net.IP {
	if ipv6 {
		return p.IP(p.counterDoTIPv6, server.IPv6)
	}
	return p.IP(p.counterDoTIPv4, server.IPv4)
}

func (p *picker) DNSServer(servers []provider.DNSServer) provider.DNSServer {
	index := 0
	if nServers := len(servers); nServers > 1 {
		index = p.randIndex(p.counterDNSServer, nServers)
	}
	return servers[index]
}

func (p *picker) DNSIP(server provider.DNSServer, ipv6 bool) net.IP {
	if ipv6 {
		return p.IP(p.counterDNSIPv6, server.IPv6)
	}
	return p.IP(p.counterDNSIPv4, server.IPv4)
}
