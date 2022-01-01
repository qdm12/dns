package nameserver

import (
	"context"
	"fmt"
	"net"
	"time"
)

type SettingsInternalDNS struct {
	// IP is the IP address to use for the DNS.
	// It defaults to 127.0.0.1 if nil.
	IP net.IP
	// Port is the port to reach the DNS server on.
	// It defaults to 53 if left unset.
	Port uint16
	// Timeout is the dialer timeout. By default there is
	// no timeout.
	Timeout time.Duration
}

func (s *SettingsInternalDNS) SetDefaults() {
	if s.IP == nil {
		s.IP = net.IPv4(127, 0, 0, 1) //nolint:gomnd
	}
	if s.Port == 0 {
		s.Port = 53
	}
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
	net.DefaultResolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialer.DialContext(ctx, "udp", net.JoinHostPort(settings.IP.String(), fmt.Sprint(settings.Port)))
	}
}
