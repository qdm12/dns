package support

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"
)

func IPv6(ctx context.Context, ipv6AddrPort netip.AddrPort) (
	ipv6Supported bool, err error) {
	if !ipv6AddrPort.IsValid() {
		const cloudflareIPv6AddrPort = "[2606:4700:4700::1111]:443"
		ipv6AddrPort = netip.MustParseAddrPort(cloudflareIPv6AddrPort)
	}

	dialer := net.Dialer{
		Timeout: time.Second,
	}
	conn, err := dialer.DialContext(ctx, "tcp", ipv6AddrPort.String())
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return false, ctxErr
		}
		if strings.HasSuffix(err.Error(), "cannot assign requested address") {
			return false, nil
		}
		return false, fmt.Errorf("unknown error: %w", err)
	}

	err = conn.Close()
	if err != nil {
		return false, fmt.Errorf("closing connection: %w", err)
	}

	return true, nil
}
