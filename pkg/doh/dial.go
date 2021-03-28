package doh

import (
	"bytes"
	"context"
	"net"
	"sync"

	"github.com/qdm12/dns/pkg/dot"
)

type dialFunc func(ctx context.Context, _, _ string) (net.Conn, error)

func newDoHDial(settings settings) dialFunc {
	// DoT HTTP client to resolve the DoH URL hostname
	dotOptions := []dot.Option{
		dot.Providers(settings.providers[0], settings.providers[1:]...),
		dot.Timeout(settings.timeout), // http client timeout really
	}
	if settings.ipv6 {
		dotOptions = append(dotOptions, dot.IPv6())
	}
	dotClient := newDoTClient(dotOptions...)

	// HTTP bodies buffer pool
	bufferPool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}

	picker := newPicker() // fast thread safe random picker

	return func(ctx context.Context, _, _ string) (conn net.Conn, err error) {
		// Pick DoH server pseudo-randomly from the chosen providers
		DoHServer := picker.DoHServer(settings.dohServers)
		// Create connection object (no actual IO yet)
		conn = newDoHConn(ctx, dotClient, bufferPool, DoHServer.URL)
		return conn, nil
	}
}
