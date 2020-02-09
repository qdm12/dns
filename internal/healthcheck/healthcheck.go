package healthcheck

import (
	"context"
	"fmt"
	"net"
)

func Healthcheck() error {
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", "127.0.0.1:53")
		},
	}
	_, err := net.LookupIP("github.com")
	if err != nil {
		return fmt.Errorf("cannot resolve github.com: %w", err)
	}
	return nil
}
