package doh

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/qdm12/dns/internal/picker"
	"github.com/qdm12/dns/internal/server"
	"github.com/qdm12/dns/pkg/dot"
	"github.com/qdm12/dns/pkg/provider"
)

func newDoHDial(settings ResolverSettings) (
	dial server.Dial, err error) {
	// note: settings are already defaulted
	metrics := settings.Metrics

	dohServers := make([]provider.DoHServer, len(settings.DoHProviders))
	for i, s := range settings.DoHProviders {
		provider, err := provider.Parse(s)
		if err != nil {
			return nil, err
		}
		dohServers[i] = provider.DoH()
	}

	// DoT HTTP client to resolve the DoH URL hostname
	DoTSettings := dot.ResolverSettings{
		DoTProviders: settings.SelfDNS.DoTProviders,
		DNSProviders: settings.SelfDNS.DNSProviders,
		Timeout:      settings.Timeout, // http client timeout really
		IPv6:         settings.SelfDNS.IPv6,
		Warner:       settings.Warner,
		Metrics:      metrics,
	}
	dotClient, err := newDoTClient(DoTSettings)
	if err != nil {
		return nil, fmt.Errorf("cannot create DoT client: %w", err)
	}

	// HTTP bodies buffer pool
	bufferPool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}

	picker := picker.New()

	return func(ctx context.Context, _, _ string) (conn net.Conn, err error) {
		// Pick DoH server pseudo-randomly from the chosen providers
		DoHServer := picker.DoHServer(dohServers)

		metrics.DoHDialInc(DoHServer.URL.String())

		// Create connection object (no actual IO yet)
		conn = newDoHConn(ctx, dotClient, bufferPool, DoHServer.URL)
		return conn, nil
	}, nil
}
