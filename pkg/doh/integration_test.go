// +build integration

package doh

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/qdm12/golibs/logging/mock_logging"
	"github.com/stretchr/testify/assert"
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
	ctrl := gomock.NewController(t)

	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})

	logger := mock_logging.NewMockLogger(ctrl)
	logger.EXPECT().Info("DNS server listening on :53")
	logger.EXPECT().Warn("DNS server stopped")

	server := NewServer(ctx, logger, Settings{})

	go server.Run(ctx, stopped)

	resolver := &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial: func(ctx context.Context, network string, address string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: time.Second}
			return dialer.DialContext(ctx, "udp", "127.0.0.1:53")
		},
	}

	const parallelResolutions = 1
	startWg := new(sync.WaitGroup)
	endWg := new(sync.WaitGroup)
	startWg.Add(parallelResolutions)
	endWg.Add(parallelResolutions)
	hostnames := []string{
		"google.com", "google.com", "github.com", "amazon.com", "cloudflare.com",
	}

	for i := 0; i < parallelResolutions; i++ {
		hostnameIndex := i % len(hostnames)
		hostname := hostnames[hostnameIndex]
		go func() {
			startWg.Done()
			startWg.Wait()
			ips, err := resolver.LookupIPAddr(ctx, hostname)
			assert.NoError(t, err)
			assert.NotEmpty(t, ips)
			t.Log(ips)
			endWg.Done()
		}()
	}

	endWg.Wait()
	cancel()
	<-stopped
}
