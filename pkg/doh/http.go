package doh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/qdm12/dns/pkg/dot"
)

type HTTPHandler interface {
	Request(ctx context.Context, url *url.URL, wire []byte) (
		respWire []byte, err error)
}

type httpHandler struct {
	client     *http.Client
	bufferPool *sync.Pool
}

func newHTTPHandler(options ...dot.Option) HTTPHandler {
	return &httpHandler{
		client: newDoTClient(options...),
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(nil)
			},
		},
	}
}

func newDoTClient(options ...dot.Option) *http.Client {
	httpTransport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{
		Resolver: dot.NewResolver(options...),
	}
	httpTransport.DialContext = dialer.DialContext
	const timeout = 5 * time.Second
	return &http.Client{
		Timeout:   timeout,
		Transport: httpTransport,
	}
}

var (
	ErrHTTPStatus = errors.New("bad HTTP status")
)

func (h *httpHandler) Request(ctx context.Context, url *url.URL, wire []byte) (
	respWire []byte, err error) {
	buffer := h.bufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer h.bufferPool.Put(buffer)

	_, err = buffer.Write(wire)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), buffer)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/dns-udpwireformat")

	response, err := h.client.Do(request)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrHTTPStatus, response.Status)
	}

	respWire, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if err := response.Body.Close(); err != nil {
		return nil, err
	}

	return respWire, nil
}
