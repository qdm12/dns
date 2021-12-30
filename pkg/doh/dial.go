package doh

import (
	"bytes"
	"context"
	"net"
	"sync"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/internal/picker"
	"github.com/qdm12/dns/pkg/dot"
	"github.com/qdm12/dns/pkg/provider"
)

type dialFunc func(ctx context.Context, _, _ string) (net.Conn, error)

type dialDNSFunc func(ctx context.Context) (*dns.Conn, error)

func wrapDial(dial dialFunc) dialDNSFunc {
	return func(ctx context.Context) (*dns.Conn, error) {
		netConn, err := dial(ctx, "", "")
		if err != nil {
			return nil, err
		}
		return &dns.Conn{Conn: netConn}, nil
	}
}

func newDoHDial(settings ResolverSettings) dialFunc {
	// note: settings are already defaulted
	metrics := settings.Metrics

	dohServers := make([]provider.DoHServer, len(settings.DoHProviders))
	for i := range settings.DoHProviders {
		dohServers[i] = settings.DoHProviders[i].DoH()
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
	dotClient := newDoTClient(DoTSettings)

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
	}
}
