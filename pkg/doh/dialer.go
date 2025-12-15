package doh

import (
	"bytes"
	"context"
	"fmt"
	"hash/maphash"
	"math/rand/v2"
	"net"
	"net/http"
	"slices"
	"sync"

	"github.com/qdm12/dns/v2/pkg/provider"
)

type Dialer struct {
	urlsSet    map[string]struct{}
	urls       []string
	servers    []provider.DoHServer
	httpClient *http.Client
	// HTTP bodies buffer pool
	bufferPool *sync.Pool
	metrics    Metrics
	randGen    *rand.Rand
}

func New(settings Settings) (dial *Dialer, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	urlsSet := make(map[string]struct{}, len(settings.UpstreamResolvers))
	urls := make([]string, len(settings.UpstreamResolvers))
	servers := make([]provider.DoHServer, len(settings.UpstreamResolvers))
	for i, upstreamResolver := range settings.UpstreamResolvers {
		urlsSet[upstreamResolver.DoH.URL] = struct{}{}
		urls[i] = upstreamResolver.DoH.URL
		servers[i] = upstreamResolver.DoH
	}

	return &Dialer{
		urlsSet:    urlsSet,
		urls:       urls,
		servers:    servers,
		httpClient: newHTTPClient(servers, settings.IPVersion),
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(nil)
			},
		},
		metrics: settings.Metrics,
		randGen: rand.New(&mapHashSource{}), //nolint:gosec
	}, nil
}

func (d *Dialer) String() string {
	return "https"
}

// ReusableConnsSupported returns true to indicate that connections created
// by this dialer can be reused.
func (d *Dialer) ReusableConnsSupported() bool {
	return false
}

func (d *Dialer) Addresses() []string {
	return slices.Clone(d.urls)
}

func (d *Dialer) Dial(ctx context.Context, _, address string) (
	conn net.Conn, err error,
) {
	url := d.pickURL(address)

	d.metrics.DoHDialInc(url)

	// Create connection object (no actual IO yet)
	conn = newDoHConn(ctx, d.httpClient, d.bufferPool, url)
	return conn, nil
}

func (d *Dialer) pickURL(address string) string {
	_, ok := d.urlsSet[address]
	if ok {
		return address
	}
	// when used as the dialer for a Go resolver, the address is the DNS system IP address
	// so pick a server url at random.
	return d.urls[d.randGen.IntN(len(d.urls))]
}

type mapHashSource struct{}

func (s *mapHashSource) Uint64() uint64 {
	return new(maphash.Hash).Sum64()
}
