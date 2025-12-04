package pool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testDialer struct {
	port uint16
}

func (d *testDialer) Addresses() []string {
	return []string{
		fmt.Sprintf("127.0.0.1:%d", d.port),
	}
}

func (d *testDialer) Dial(ctx context.Context, network, address string,
) (net.Conn, error) {
	dialer := &net.Dialer{}
	return dialer.DialContext(ctx, network, address)
}

func startLocalTCPServer(t *testing.T) (
	dialer *testDialer, runError <-chan error,
) {
	t.Helper()
	listener, err := net.ListenTCP("tcp", nil)
	require.NoError(t, err)

	runErrorCh := make(chan error)
	runError = runErrorCh
	done := make(chan struct{})

	ready := make(chan struct{})
	go func() {
		defer close(done)
		close(ready)
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				runErrorCh <- fmt.Errorf("accepting connection: %w", err)
				return
			}
			go handleConnection(conn, runErrorCh)
		}
	}()

	stop := func() {
		_ = listener.Close()
		<-done
	}
	t.Cleanup(stop)

	select {
	case <-ready:
	case err := <-runError:
		t.Fatal("server failed to start:", err)
	}

	port := uint16(listener.Addr().(*net.TCPAddr).Port) //nolint:gosec,forcetypeassert
	return &testDialer{port: port}, runError
}

func handleConnection(conn net.Conn, runErrorCh chan<- error) {
	defer conn.Close()
	const timeout = time.Minute
	err := conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		runErrorCh <- fmt.Errorf("setting deadline: %w", err)
		return
	}
	_, err = io.Copy(conn, conn)
	if err != nil {
		runErrorCh <- fmt.Errorf("copying: %w", err)
		return
	}
}

// checkConnWorks checks the connection works with the echo server with behavior
// defined in [handleConnection]. It also closes the connection when done.
func checkConnWorks(t *testing.T, conn net.Conn) {
	t.Helper()
	require.NotNil(t, conn)
	message := []byte("hello")
	_, err := conn.Write(message)
	require.NoError(t, err)

	reply := make([]byte, len(message))
	_, err = io.ReadFull(conn, reply)
	require.NoError(t, err)
	assert.Equal(t, message, reply)

	err = conn.Close()
	assert.NoError(t, err)
}

type noopConn struct {
	net.Conn
}

func (*noopConn) Close() error {
	return nil
}

func setFieldsForAddrConns(addrConns []addressConns) {
	for i := range addrConns {
		for j := range addrConns[i].conns {
			addrConns[i].conns[j].addrIndex = i
			addrConns[i].conns[j].connIndex = j
			addrConns[i].conns[j].Conn = &noopConn{}
		}
	}
}

func clearPoolFieldsForComparison(p *Pool) {
	p.timeNow = nil
	p.metrics = nil
}
