package doh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/qdm12/dns/v2/pkg/dot"
)

var (
	ErrHTTPStatus = errors.New("bad HTTP status")
)

func newDoTClient(settings dot.ResolverSettings) (
	client *http.Client, err error) {
	httpTransport := http.DefaultTransport.(*http.Transport).Clone() //nolint:forcetypeassert

	resolver, err := dot.NewResolver(settings)
	if err != nil {
		return nil, fmt.Errorf("cannot create DoT resolver: %w", err)
	}

	dialer := &net.Dialer{
		Resolver: resolver,
	}
	httpTransport.DialContext = dialer.DialContext
	const timeout = 5 * time.Second
	return &http.Client{
		Timeout:   timeout,
		Transport: httpTransport,
	}, nil
}

func dohHTTPRequest(ctx context.Context, client *http.Client, bufferPool *sync.Pool,
	url *url.URL, wire []byte) (respWire []byte, err error) { //nolint:interfacer
	buffer := bufferPool.Get().(*bytes.Buffer) //nolint:forcetypeassert
	buffer.Reset()
	defer bufferPool.Put(buffer)

	_, err = buffer.Write(wire)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), buffer)
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
