package dot

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"

	"github.com/qdm12/dns/pkg/provider"
)

type dialFunc func(ctx context.Context, _, _ string) (net.Conn, error)

func newDoTDial(settings ResolverSettings) dialFunc {
	dotServers := make([]provider.DoTServer, len(settings.DoTProviders))
	for i := range settings.DoTProviders {
		dotServers[i] = settings.DoTProviders[i].DoT()
	}

	dnsServers := make([]provider.DNSServer, len(settings.DNSProviders))
	for i := range settings.DNSProviders {
		dnsServers[i] = settings.DNSProviders[i].DNS()
	}

	dialer := &net.Dialer{
		Timeout: settings.Timeout,
	}

	picker := newPicker()

	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		DoTServer := picker.DoTServer(dotServers)
		ip := picker.DoTIP(DoTServer, settings.IPv6)
		tlsAddr := net.JoinHostPort(ip.String(), strconv.Itoa(int(DoTServer.Port)))

		conn, err := dialer.DialContext(ctx, "tcp", tlsAddr)
		if err != nil {
			if len(dnsServers) > 0 {
				// fallback on plain DNS if DoT does not work
				dnsServer := picker.DNSServer(dnsServers)
				ip := picker.DNSIP(dnsServer, settings.IPv6)
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
	}
}
