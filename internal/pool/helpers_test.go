package pool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"

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

func startLocalTCPServer(t *testing.T, handleConn func(net.Conn) error) ( //nolint:cyclop
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
				if errors.Is(err, net.ErrClosed) {
					return
				}
				runErrorCh <- fmt.Errorf("accepting connection: %w", err)
				return
			}
			connsInFlightMutex.Lock()
			connsInFlight[conn.RemoteAddr().String()] = conn
			connsInFlightMutex.Unlock()
			handleConnWg.Add(1)
			handleConnWg.Go(func() {
				defer handleConnWg.Done()
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

	t.Cleanup(func() {
		_ = listener.Close()
		// drain error channel in case test exited with fatal and did not read the runError
		// channel return, and one or more goroutines are trying to write an error to runErrorCh
		for range len(connsInFlight) {
			select {
			case <-runErrorCh:
			default:
			}
		}
		<-listenerDone
		connsInFlightMutex.Lock()
		for _, conn := range connsInFlight {
			_ = conn.Close()
		}
		connsInFlightMutex.Unlock()
		handleConnWg.Wait()
	})

	select {
	case <-ready:
	case err := <-runError:
		t.Fatal("server failed to start:", err)
	}

	port := uint16(listener.Addr().(*net.TCPAddr).Port) //nolint:gosec,forcetypeassert
	return &testDialer{port: port}, runError
}

// handleConnCopy echoes back the 4 bytes of data received.
func handleConnCopy(conn net.Conn) error {
	const length = 4
	buffer := make([]byte, length)
	for {
		_, err := io.CopyBuffer(conn, conn, buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("copying: %w", err)
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
	p.nextConnID = 0
	normalizeAddrConnsForComparison(p.addrConns)
}

func normalizeAddrConnsForComparison(addrConns []addressConns) {
	for i := range addrConns {
		addrConns[i].connIDToIndex = nil
		for j := range addrConns[i].conns {
			addrConns[i].conns[j].id = 0
		}
	}
}
