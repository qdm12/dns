package dot

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/qdm12/dns/v2/internal/picker"
	"github.com/qdm12/dns/v2/internal/server"
	"github.com/qdm12/dns/v2/pkg/provider"
)

func newDoTDial(settings ResolverSettings) (dial server.Dial) {
	warner := settings.Warner
	metrics := settings.Metrics

	dotServers := make([]provider.DoTServer, len(settings.UpstreamResolvers))
	for i, upstreamResolver := range settings.UpstreamResolvers {
		dotServers[i] = upstreamResolver.DoT
	}

	dialer := &net.Dialer{
		Timeout: settings.Timeout,
	}

	picker := picker.New()
	ipv6 := settings.IPVersion == "ipv6"

	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		serverName, serverAddress := pickNameAddress(picker,
			dotServers, ipv6)

		conn, err := dialer.DialContext(ctx, "tcp", serverAddress)
		if err != nil {
			warner.Warn(err.Error())
			metrics.DoTDialInc(serverName, serverAddress, "error")
			return nil, err
		}

		metrics.DoTDialInc(serverName, serverAddress, "success")

		tlsConf := &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: serverName,
		}
		// TODO handshake? See tls.DialWithDialer
		return tls.Client(conn, tlsConf), nil
	}
}

func pickNameAddress(picker *picker.Picker, servers []provider.DoTServer,
	ipv6 bool) (name, address string) {
	server := picker.DoTServer(servers)
	addrPort := picker.DoTAddrPort(server, ipv6)
	return server.Name, addrPort.String()
}
