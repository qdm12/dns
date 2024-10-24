package plain

import (
	"context"
	"fmt"
	"net"

	"github.com/qdm12/dns/v2/internal/picker"
	"github.com/qdm12/dns/v2/pkg/provider"
)

type Dialer struct {
	picker    *picker.Picker
	servers   []provider.PlainServer
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

	servers := make([]provider.PlainServer, len(settings.UpstreamResolvers))
	for i, upstreamResolver := range settings.UpstreamResolvers {
		servers[i] = upstreamResolver.Plain
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
	return "dns over plaintext"
}

func (d *Dialer) Dial(ctx context.Context, network, _ string) (
	conn net.Conn, err error,
) {
	serverAddress := pickAddress(d.picker, d.servers, d.ipv6)

	udpConn, err := d.netDialer.DialContext(ctx, network, serverAddress)
	if err != nil {
		d.warner.Warn(err.Error())
		d.metrics.PlainDialInc(serverAddress, "error")
		return nil, err
	}

	d.metrics.PlainDialInc(serverAddress, "success")
	return udpConn, nil
}

func pickAddress(picker *picker.Picker, servers []provider.PlainServer,
	ipv6 bool,
) (address string) {
	server := picker.PlainServer(servers)
	addrPort := picker.PlainAddrPort(server, ipv6)
	return addrPort.String()
}
