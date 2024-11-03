package doh

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/qdm12/dns/v2/internal/picker"
	"github.com/qdm12/dns/v2/pkg/provider"
)

type Dialer struct {
	picker     *picker.Picker
	servers    []provider.DoHServer
	httpClient *http.Client
	// HTTP bodies buffer pool
	bufferPool *sync.Pool
	metrics    Metrics
}

func New(settings Settings) (dial *Dialer, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	servers := make([]provider.DoHServer, len(settings.UpstreamResolvers))
	for i, upstreamResolver := range settings.UpstreamResolvers {
		servers[i] = upstreamResolver.DoH
	}

	return &Dialer{
		picker:     picker.New(),
		servers:    servers,
		httpClient: newHTTPClient(servers, settings.IPVersion),
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(nil)
			},
		},
		metrics: settings.Metrics,
	}, nil
}

func (d *Dialer) String() string {
	return "https"
}

func (d *Dialer) Dial(ctx context.Context, _, _ string) (
	conn net.Conn, err error,
) {
	// Pick DoH server pseudo-randomly from the chosen providers
	DoHServer := d.picker.DoHServer(d.servers)

	d.metrics.DoHDialInc(DoHServer.URL)

	// Create connection object (no actual IO yet)
	conn = newDoHConn(ctx, d.httpClient, d.bufferPool, DoHServer.URL)
	return conn, nil
}
