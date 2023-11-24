package doh

import (
	"bytes"
	"context"
	"net"
	"sync"

	"github.com/qdm12/dns/v2/internal/server"
	"github.com/qdm12/dns/v2/pkg/provider"
)

func newDoHDial(settings ResolverSettings) (dial server.Dial) {
	// note: settings are already defaulted
	metrics := settings.Metrics

	dohServers := make([]provider.DoHServer, len(settings.DoHProviders))
	for i, provider := range settings.DoHProviders {
		dohServers[i] = provider.DoH
	}

	httpClient := newHTTPClient(dohServers, settings.IPVersion)

	// HTTP bodies buffer pool
	bufferPool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}

	picker := settings.Picker

	return func(ctx context.Context, _, _ string) (conn net.Conn, err error) {
		// Pick DoH server pseudo-randomly from the chosen providers
		DoHServer := picker.DoHServer(dohServers)

		metrics.DoHDialInc(DoHServer.URL)

		// Create connection object (no actual IO yet)
		conn = newDoHConn(ctx, httpClient, bufferPool, DoHServer.URL)
		return conn, nil
	}
}
