package dns

import (
	"context"
	"net"
)

// UseDNSInternally is to change the Go program DNS only
func (c *configurator) UseDNSInternally(ip net.IP) {
	c.logger.Info("using DNS address %s internally", ip.String())
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", net.JoinHostPort(ip.String(), "53"))
		},
	}
}
