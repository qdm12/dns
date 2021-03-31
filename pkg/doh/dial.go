package doh

import (
	"bytes"
	"context"
	"net"
	"sync"

	"github.com/qdm12/dns/pkg/dot"
)

type dialFunc func(ctx context.Context, _, _ string) (net.Conn, error)

func newDoHDial(settings Settings) dialFunc {
	// DoT HTTP client to resolve the DoH URL hostname
	DoTSettings := dot.Settings{
		DoTServers: settings.SelfDNS.DoTServers,
		DNSServers: settings.SelfDNS.DNSServers,
		Timeout:    settings.Timeout, // http client timeout really
		IPv6:       settings.IPv6,
	}
	dotClient := newDoTClient(DoTSettings)

	// HTTP bodies buffer pool
	bufferPool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}

	picker := newPicker() // fast thread safe random picker

	return func(ctx context.Context, _, _ string) (conn net.Conn, err error) {
		// Pick DoH server pseudo-randomly from the chosen providers
		DoHServer := picker.DoHServer(settings.DoHServers)
		// Create connection object (no actual IO yet)
		conn = newDoHConn(ctx, dotClient, bufferPool, DoHServer.URL)
		return conn, nil
	}
}
