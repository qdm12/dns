// +build integration

package dot

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/qdm12/golibs/logging"
	"github.com/stretchr/testify/require"
)

func Test_Resolver(t *testing.T) {
	t.Parallel()

	const hostname = "google.com"

	resolver := NewResolver(Settings{})

	ips, err := resolver.LookupIPAddr(context.Background(), hostname)

	require.NoError(t, err)
	require.NotEmpty(t, ips)
	t.Logf("resolved %s to: %v", hostname, ips)
}

func Test_Server(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})

	logger := logging.New(logging.StdLog)
	server := NewServer(ctx, logger, Settings{})

	go server.Run(ctx, stopped)

	const hostname = "google.com" // we use google.com as github.com doesn't have an IPv6 :(
	resolver := &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial: func(ctx context.Context, network string, address string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: time.Second}
			return dialer.DialContext(ctx, "udp", "127.0.0.1:53")
		},
	}

	ips, err := resolver.LookupIPAddr(ctx, hostname)

	require.NoError(t, err)
	require.NotEmpty(t, ips)
	t.Logf("resolved %s to: %v", hostname, ips)

	cancel()
	<-stopped
}
