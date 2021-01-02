package dns

import (
	"context"
	"net"
)

// UseDNSInternally is to change the Go program DNS only.
func (c *configurator) UseDNSInternally(ip net.IP) {
	net.DefaultResolver.PreferGo = true
	net.DefaultResolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{}
		return d.DialContext(ctx, "udp", net.JoinHostPort(ip.String(), "53"))
	}
}
