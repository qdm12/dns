// Package health provides health check functionality, including a health check
// server and a client to query the health status.
package health

import (
	"context"
	"fmt"
	"net"
)

// IsHealthy checks the localhost DNS UDP server is working by
// resolving github.com.
func IsHealthy(ctx context.Context) (err error) {
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", "127.0.0.1:53")
		},
	}
	_, err = net.DefaultResolver.LookupIPAddr(ctx, "github.com")
	if err != nil {
		return fmt.Errorf("resolving github.com: %w", err)
	}
	return nil
}
