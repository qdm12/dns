package pool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
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

func startLocalTCPServer(t *testing.T, handleConn func(net.Conn) error) (
	dialer *testDialer, runError <-chan error,
) {
	t.Helper()
	listener, err := net.ListenTCP("tcp", nil)
	require.NoError(t, err)

	runErrorCh := make(chan error)
	runError = runErrorCh
	listenerDone := make(chan struct{})
	var handleConnWg sync.WaitGroup

	connsInFlight := make(map[string]net.Conn)
	var connsInFlightMutex sync.Mutex

	ready := make(chan struct{})
	go func() {
		defer close(listenerDone)
		close(ready)
		for {
			conn, err := listener.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					sendError(runErrorCh, fmt.Errorf("accepting connection: %w", err))
				}
				return
			}
			// Registered before the handler goroutine is accounted for, so stop()
			// cannot sweep the map and miss a connection it then never closes, leaving
			// handleConnWg.Wait blocked on a handler reading a live connection.
			connsInFlightMutex.Lock()
			connsInFlight[conn.RemoteAddr().String()] = conn
			connsInFlightMutex.Unlock()
			handleConnWg.Go(func() {
				err := handleConn(conn)
				if err != nil {
					select {
					case <-listenerDone: // server stopped
					case runErrorCh <- err:
					}
				}
			})
		}
	}()

	stop := func() {
		_ = listener.Close()
		<-listenerDone
		connsInFlightMutex.Lock()
		for _, conn := range connsInFlight {
			_ = conn.Close()
		}
		connsInFlightMutex.Unlock()
		handleConnWg.Wait()
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

// sendError reports err without ever blocking, since the test may have stopped
// reading from the channel; a blocking send here would prevent the accept loop
// from returning and listenerDone from closing.
func sendError(runErrorCh chan<- error, err error) {
	select {
	case runErrorCh <- err:
	default:
	}
}

// handleConnCopy echoes back the data received, until the peer closes the
// connection.
//
// It reads and writes explicitly instead of calling [io.CopyBuffer] with the
// same connection as source and destination: net then splices the connection
// onto itself (same file descriptor in and out), and poll.Splice reports the
// peer EOF as (0, nil) instead of [io.EOF], which makes the copy loop spin
// forever on a closed connection instead of returning.
func handleConnCopy(conn net.Conn) error {
	const length = 4
	buffer := make([]byte, length)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("reading: %w", err)
		}
		if _, err := conn.Write(buffer[:n]); err != nil {
			return fmt.Errorf("writing: %w", err)
		}
	}
}

// checkConnCopies checks the connection works with the echo server with behavior
// defined in [handleConnCopy]. It also closes the connection when done.
func checkConnCopies(t *testing.T, conn net.Conn) {
	t.Helper()
	require.NotNil(t, conn)
	const length = 4
	message := make([]byte, length)
	for i := range message {
		message[i] = byte(i)
	}
	_, err := conn.Write(message)
	require.NoError(t, err)

	reply := make([]byte, len(message))
	_, err = io.ReadFull(conn, reply)
	require.NoError(t, err)
	assert.Equal(t, message, reply)

	err = conn.Close()
	assert.NoError(t, err)
}

// Test_handleConnCopy checks the echo handler terminates once the peer closes
// the connection, instead of spinning forever on the closed connection.
func Test_handleConnCopy(t *testing.T) {
	t.Parallel()

	listener, err := net.ListenTCP("tcp", nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	dialer := &testDialer{port: uint16(listener.Addr().(*net.TCPAddr).Port)} //nolint:gosec,forcetypeassert

	handleConnError := make(chan error, 1)
	go func() {
		serverConn, err := listener.Accept()
		if err != nil {
			handleConnError <- fmt.Errorf("accepting: %w", err)
			return
		}
		handleConnError <- handleConnCopy(serverConn)
	}()

	netConn, err := dialer.Dial(context.Background(), "tcp", dialer.Addresses()[0])
	require.NoError(t, err)
	// Closes the connection once the 4 bytes have been echoed back.
	checkConnCopies(t, netConn)

	// Bounded well below the test timeout, so a handler spinning on a closed
	// connection fails this test instead of hanging the whole test suite.
	const maxWait = 10 * time.Second
	wait := maxWait
	if deadline, ok := t.Deadline(); ok {
		if remaining := time.Until(deadline); remaining < wait {
			wait = remaining
		}
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case err := <-handleConnError:
		require.NoError(t, err)
	case <-timer.C:
		t.Fatal("handleConnCopy did not return after the peer closed the connection")
	}
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
			addrConns[i].conns[j].addrIndex = i      //nolint:gosec
			addrConns[i].conns[j].connIndex = j      //nolint:gosec
			addrConns[i].conns[j].Conn = &noopConn{} //nolint:gosec
		}
	}
}

func clearPoolFieldsForComparison(p *Pool) {
	p.timeNow = nil
	p.metrics = nil
	p.nextConnID = 0
	normalizeAddrConnsForComparison(p.addrConns)
}

func normalizeAddrConnsForComparison(addrConns []addressConns) {
	for i := range addrConns {
		addrConns[i].connIDToIndex = nil
		for j := range addrConns[i].conns {
			addrConns[i].conns[j].id = 0 //nolint:gosec
		}
	}
}
