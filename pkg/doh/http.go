package doh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/qdm12/dns/v2/pkg/provider"
)

var ErrHTTPStatus = errors.New("bad HTTP status")

func newHTTPClient(dohServers []provider.DoHServer, ipVersion string) (
	client *http.Client,
) {
	httpTransport := http.DefaultTransport.(*http.Transport).Clone() //nolint:forcetypeassert

	dialer := &net.Dialer{
		Resolver: newHTTPClientResolver(dohServers, ipVersion),
	}
	httpTransport.DialContext = dialer.DialContext
	const timeout = 5 * time.Second
	return &http.Client{
		Timeout:   timeout,
		Transport: httpTransport,
	}
}

func newHTTPClientResolver(dohServers []provider.DoHServer,
	ipVersion string,
) *net.Resolver {
	// Compute mappings early and once for all dial calls
	fqdnToIPv4, fqdnToIPv6 := dohServersToHardcodedMaps(dohServers, ipVersion)

	return &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial: func(_ context.Context, _, _ string) (net.Conn, error) {
			return &hardcodedConn{
				fqdnToIPv4: fqdnToIPv4,
				fqdnToIPv6: fqdnToIPv6,
			}, nil
		},
	}
}

func dohHTTPRequest(ctx context.Context, client *http.Client, bufferPool *sync.Pool,
	url string, wire []byte,
) (respWire []byte, err error) { //nolint:interfacer
	buffer := bufferPool.Get().(*bytes.Buffer) //nolint:forcetypeassert
	buffer.Reset()
	defer bufferPool.Put(buffer)

	_, err = buffer.Write(wire)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buffer)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/dns-message")
	request.Header.Set("Accept", "application/dns-message")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrHTTPStatus, response.Status)
	}

	respWire, err = io.ReadAll(response.Body) // TODO copy to buffer
	if err != nil {
		return nil, err
	}

	if err := response.Body.Close(); err != nil {
		return nil, err
	}

	return respWire, nil
}
