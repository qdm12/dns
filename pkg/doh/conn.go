package doh

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func newDoHConn(ctx context.Context, client *http.Client,
	bufferPool *sync.Pool, dohURL *url.URL) net.Conn {
	ctx, cancel := context.WithCancel(ctx)
	const maxUDPSize = 4096
	return &dohConn{
		ctx:        ctx,
		client:     client,
		bufferPool: bufferPool,
		dohURL:     dohURL,
		inBuffer:   bytes.NewBuffer(make([]byte, 0, maxUDPSize)),
		outBuffer:  bytes.NewBuffer(make([]byte, 0, maxUDPSize)),
		cancel:     cancel,
	}
}

type dohConn struct {
	// External objects injected at creation
	ctx        context.Context
	client     *http.Client
	bufferPool *sync.Pool
	dohURL     *url.URL

	// Internals
	inBuffer  *bytes.Buffer // TODO obtain from syncPool
	outBuffer *bytes.Buffer // TODO obtain from syncPool
	cancel    context.CancelFunc
	deadline  time.Time
}

func (c *dohConn) readOutputBuffer(b []byte) (n int, err error) {
	if c.outBuffer.Len() == 0 {
		return 0, nil
	}
	return c.outBuffer.Read(b)
}

func (c *dohConn) Read(b []byte) (n int, err error) {
	// TODO move http request in Write perhaps?
	n, err = c.readOutputBuffer(b)
	if err != nil {
		return n, err
	} else if n > 0 {
		// We had the result of a previous HTTP request
		// to the DoH server, so return here.
		return n, nil
	}

	// The output buffer is empty, so this is a fresh request we need
	// to execute against the DoH server. This happens only once
	// on the first read for the connection.

	dnsQueryBytes := c.inBuffer.Bytes()
	dnsQueryBytes = dnsQueryBytes[2:]

	c.ctx, c.cancel = context.WithCancel(c.ctx)
	c.ctx, c.cancel = context.WithDeadline(c.ctx, c.deadline)

	dnsAnswerBytes, err := dohHTTPRequest(c.ctx, c.client, c.bufferPool, c.dohURL, dnsQueryBytes)
	c.cancel()
	if err != nil {
		return 0, err
	}

	if err := c.writeToOutputBuffer(dnsAnswerBytes); err != nil {
		return 0, err
	}

	return c.outBuffer.Read(b)
}

// Write only writes the bytes to send in the connection
// to a buffer. The HTTP request is made in Read instead
// such that response data can be read at the same time.
func (c *dohConn) Write(b []byte) (n int, err error) {
	return c.inBuffer.Write(b)
}

func (c *dohConn) Close() error {
	c.cancel()
	// TODO put back in sync.Pool here?
	return nil
}

func (c *dohConn) LocalAddr() net.Addr {
	return nil
}

func (c *dohConn) RemoteAddr() net.Addr {
	return nil
}

func (c *dohConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

func (c *dohConn) SetReadDeadline(t time.Time) error {
	c.deadline = t
	return nil
}

func (c *dohConn) SetWriteDeadline(t time.Time) error {
	// IO happens in read only so no timeout to set here
	return nil
}

func (c *dohConn) writeToOutputBuffer(b []byte) (err error) {
	// Write the size of the message in the first two bytes
	const bitsInByte = 8
	if err := c.outBuffer.WriteByte(byte(len(b) >> bitsInByte)); err != nil {
		return err
	}

	if err := c.outBuffer.WriteByte(byte(len(b))); err != nil {
		return err
	}

	_, err = c.outBuffer.Write(b)
	return err
}
