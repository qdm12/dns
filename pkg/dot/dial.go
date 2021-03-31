package dot

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
)

type dialFunc func(ctx context.Context, _, _ string) (net.Conn, error)

func newDoTDial(settings Settings) dialFunc {
	dialer := &net.Dialer{
		Timeout: settings.Timeout,
	}

	picker := newPicker()

	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		DoTServer := picker.DoTServer(settings.DoTServers)
		ip := picker.DoTIP(DoTServer, settings.IPv6)
		tlsAddr := net.JoinHostPort(ip.String(), strconv.Itoa(int(DoTServer.Port)))

		conn, err := dialer.DialContext(ctx, "tcp", tlsAddr)
		if err != nil {
			if len(settings.DNSServers) > 0 {
				// fallback on plain DNS if DoT does not work
				dnsServer := picker.DNSServer(settings.DNSServers)
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
