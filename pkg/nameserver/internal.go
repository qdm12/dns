package nameserver

import (
	"context"
	"net"
	"net/netip"
	"time"

	"github.com/qdm12/gosettings"
)

type SettingsInternalDNS struct {
	// AddrPort is the address to use for the DNS.
	// It defaults to 127.0.0.1:53.
	AddrPort netip.AddrPort
	// Timeout is the dialer timeout. By default there is
	// no timeout.
	Timeout time.Duration
}

func (s *SettingsInternalDNS) SetDefaults() {
	const defaultPort = 53
	defaultAddrPort := netip.AddrPortFrom(netip.AddrFrom4([4]byte{127, 0, 0, 1}), defaultPort)
	s.AddrPort = gosettings.DefaultValidator(s.AddrPort, defaultAddrPort)
}

func (s SettingsInternalDNS) Validate() (err error) {
	return nil
}

// UseDNSInternally changes the Go program DNS only.
func UseDNSInternally(settings SettingsInternalDNS) {
	settings.SetDefaults()

	dialer := net.Dialer{
		Timeout: settings.Timeout,
	}

	net.DefaultResolver.PreferGo = true
	net.DefaultResolver.Dial = func(ctx context.Context, _, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, "udp", settings.AddrPort.String())
	}
}
