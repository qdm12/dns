package support

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"
)

func IPv6(ctx context.Context) (ipv6Supported bool, err error) {
	dialer := net.Dialer{
		Timeout: time.Second,
	}

	ipv6AddrPorts := []netip.AddrPort{
		netip.MustParseAddrPort("[2606:4700:4700::1111]:443"), // Cloudflare
		netip.MustParseAddrPort("[2001:4860:4860::8888]:443"), // Google
	}

	for _, ipv6AddrPort := range ipv6AddrPorts {
		conn, connErr := dialer.DialContext(ctx, "tcp", ipv6AddrPort.String())
		if connErr != nil {
			if ctx.Err() != nil {
				return false, connErr
			}
			errMessage := connErr.Error()
			ipv6ErrorMessages := []string{
				"connect: network is unreachable",
				"cannot assign requested address",
			}
			for _, ipv6ErrorMessage := range ipv6ErrorMessages {
				if strings.Contains(errMessage, ipv6ErrorMessage) {
					return false, nil
				}
			}
			err = connErr
			continue // try next IPv6 address
		}

		err = conn.Close()
		if err != nil {
			return false, fmt.Errorf("closing connection: %w", err)
		}

		return true, nil
	}
	return false, err // return last error
}
