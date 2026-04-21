package pool

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"

	noopmetrics "github.com/qdm12/dns/v2/internal/pool/metrics/noop"
	"github.com/stretchr/testify/require"
)

func Test_Pool_stress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	t.Parallel()

	dialer, runErr := startLocalTCPServer(t, handleConnCopy)
	pool := New(dialer, noopmetrics.New())

	const workers = 16
	const iterations = 100

	resultCh := make(chan error)
	start := make(chan struct{})
	for i := range workers {
		go runStressWorker(pool, i, iterations, resultCh, start)
	}

	close(start)
	var errs []error
	for range workers {
		err := <-resultCh
		if err != nil {
			errs = append(errs, err)
		}
	}
	require.Empty(t, errs)

	pool.mutex.Lock()
	currentConns := len(pool.addrConns[0].conns)
	pool.mutex.Unlock()
	require.LessOrEqual(t, currentConns, workers,
		"pool retained too many live connections after stress run")

	select {
	case err := <-runErr:
		require.NoError(t, err)
	default:
	}
}

func runStressWorker(pool *Pool, worker int, iterations int,
	resultCh chan<- error, start <-chan struct{},
) {
	<-start
	for i := range iterations {
		conn, err := pool.Get(context.Background(), "tcp")
		if err != nil {
			resultCh <- fmt.Errorf("worker %d: iteration %d: get: %w", worker, i, err)
			return
		}

		switch rand.IntN(4) { //nolint:gosec
		case 0:
			pool.Put(conn)
		case 1:
			_ = conn.Close()
			pool.PutDead(conn)
		case 2:
			renewedConn, err := pool.Renew(context.Background(), "tcp", conn)
			if err != nil {
				// Failed renew already marks the slot as dead in pool state.
				continue
			}
			pool.Put(renewedConn)
		default:
			pool.Put(conn)
		}
	}
	fmt.Println("worker", worker, "done")
	resultCh <- nil
}
