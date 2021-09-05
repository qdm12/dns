package dot

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/qdm12/dns/pkg/provider"
)

type dialFunc func(ctx context.Context, _, _ string) (net.Conn, error)

func newDoTDial(settings ResolverSettings) dialFunc {
	warner := settings.Warner
	metrics := settings.Metrics

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
		tlsAddr := net.JoinHostPort(ip.String(), fmt.Sprint(DoTServer.Port))

		conn, err := dialer.DialContext(ctx, "tcp", tlsAddr)
		if err != nil {
			warner.Warn(err.Error())

			metrics.DoTDialInc(DoTServer.Name, tlsAddr, "error")

			if len(dnsServers) > 0 {
				// fallback on plain DNS if DoT does not work
				dnsServer := picker.DNSServer(dnsServers)
				ip := picker.DNSIP(dnsServer, settings.IPv6)
				ipStr := ip.String()
				plainAddr := net.JoinHostPort(ipStr, "53")
				conn, err := dialer.DialContext(ctx, "udp", plainAddr)
				if err != nil {
					warner.Warn(err.Error())
					metrics.DNSDialInc(ipStr, plainAddr, "error")
					return conn, err
				}
				metrics.DNSDialInc(ipStr, plainAddr, "success")
				return conn, nil
			}
			return nil, err
		}

		metrics.DoTDialInc(DoTServer.Name, tlsAddr, "success")

		tlsConf := &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: DoTServer.Name,
		}
		// TODO handshake? See tls.DialWithDialer
		return tls.Client(conn, tlsConf), nil
	}
}
