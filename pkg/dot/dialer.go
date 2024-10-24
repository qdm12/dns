package dot

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/qdm12/dns/v2/internal/picker"
	"github.com/qdm12/dns/v2/pkg/provider"
)

type Dialer struct {
	picker    *picker.Picker
	servers   []provider.DoTServer
	ipv6      bool
	netDialer *net.Dialer
	warner    Warner
	metrics   Metrics
}

func New(settings Settings) (dial *Dialer, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	servers := make([]provider.DoTServer, len(settings.UpstreamResolvers))
	for i, upstreamResolver := range settings.UpstreamResolvers {
		servers[i] = upstreamResolver.DoT
	}

	return &Dialer{
		picker:  picker.New(),
		servers: servers,
		ipv6:    settings.IPVersion == "ipv6",
		netDialer: &net.Dialer{
			Timeout: settings.Timeout,
		},
		warner:  settings.Warner,
		metrics: settings.Metrics,
	}, nil
}

func (d *Dialer) String() string {
	return "dns over tls"
}

func (d *Dialer) Dial(ctx context.Context, _, _ string) (
	conn net.Conn, err error,
) {
	serverName, serverAddress := pickNameAddress(d.picker,
		d.servers, d.ipv6)

	conn, err = d.netDialer.DialContext(ctx, "tcp", serverAddress)
	if err != nil {
		d.warner.Warn(err.Error())
		d.metrics.DoTDialInc(serverName, serverAddress, "error")
		return nil, err
	}

	d.metrics.DoTDialInc(serverName, serverAddress, "success")

	tlsConf := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: serverName,
	}
	// TODO handshake? See tls.DialWithDialer
	return tls.Client(conn, tlsConf), nil
}

func pickNameAddress(picker *picker.Picker, servers []provider.DoTServer,
	ipv6 bool,
) (name, address string) {
	server := picker.DoTServer(servers)
	addrPort := picker.DoTAddrPort(server, ipv6)
	return server.Name, addrPort.String()
}
