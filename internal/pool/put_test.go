package pool

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_Pool_Put(t *testing.T) {
	t.Parallel()

	now := time.Unix(10000, 0)

	t.Run("not_a_poolConn", func(t *testing.T) {
		t.Parallel()
		pool := &Pool{}
		conn := &net.TCPConn{}
		assert.PanicsWithValue(t, "cannot put back non-pool connection *net.TCPConn", func() {
			pool.Put(conn)
		})
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		metrics := NewMockMetrics(ctrl)
		dialer := &testDialer{
			port: 853,
		}
		address := dialer.Addresses()[0] // 127.0.0.1:853

		// Put call metrics
		metrics.EXPECT().PutConnsInc(address, "live")
		// cleanup metrics
		metrics.EXPECT().DeadConnsInc(address)
		metrics.EXPECT().RemovedConnsAdd(address, uint(2))

		pool := New(dialer, metrics)
		pool.timeNow = func() time.Time { return now }
		conns := []poolConn{
			{inUse: true}, // in use and cannot be removed
			{lastUsed: now.Add(-maxIdleDuration + 1)}, // not expired yet
			{lastUsed: now.Add(-maxIdleDuration - 1)}, // expired
			{dead: true},  // marked as dead already
			{inUse: true}, // the one we put back
		}
		pool.addrConns[0].conns = conns
		setFieldsForAddrConns(pool.addrConns)

		expectedConns := []poolConn{
			{inUse: true},
			{lastUsed: now.Add(-maxIdleDuration + 1)},
			{lastUsed: now}, // the one we put back
		}
		expectedAddrConns := []addressConns{{
			address: address,
			conns:   expectedConns,
		}}
		setFieldsForAddrConns(expectedAddrConns)

		conn := conns[len(conns)-1]

		pool.Put(conn)

		assert.Equal(t, expectedAddrConns, pool.addrConns)
	})
}

func Test_Pool_PutDead(t *testing.T) {
	t.Parallel()

	now := time.Unix(10000, 0)

	t.Run("not_a_poolConn", func(t *testing.T) {
		t.Parallel()
		pool := &Pool{}
		conn := &net.TCPConn{}
		assert.PanicsWithValue(t, "cannot put back dead non-pool connection *net.TCPConn", func() {
			pool.PutDead(conn)
		})
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		metrics := NewMockMetrics(ctrl)
		dialer := &testDialer{
			port: 853,
		}
		address := dialer.Addresses()[0] // 127.0.0.1:853

		// PutDead call metrics
		metrics.EXPECT().PutConnsInc(address, "dead")
		metrics.EXPECT().DeadConnsInc(address) // connection put as dead
		// cleanup metrics
		metrics.EXPECT().DeadConnsInc(address) // another expired
		metrics.EXPECT().RemovedConnsAdd(address, uint(3))

		pool := New(dialer, metrics)
		pool.timeNow = func() time.Time { return now }
		conns := []poolConn{
			{inUse: true}, // in use and cannot be removed
			{lastUsed: now.Add(-maxIdleDuration + 1)}, // not expired yet
			{lastUsed: now.Add(-maxIdleDuration - 1)}, // expired
			{dead: true},  // marked as dead already
			{inUse: true}, // the one we put back
		}
		pool.addrConns[0].conns = conns
		setFieldsForAddrConns(pool.addrConns)

		expectedConns := []poolConn{
			{inUse: true},
			{lastUsed: now.Add(-maxIdleDuration + 1)},
		}
		expectedAddrConns := []addressConns{{
			address: address,
			conns:   expectedConns,
		}}
		setFieldsForAddrConns(expectedAddrConns)

		conn := conns[len(conns)-1]

		pool.PutDead(conn)

		assert.Equal(t, expectedAddrConns, pool.addrConns)
	})
}

func Test_Pool_cleanup(t *testing.T) {
	t.Parallel()

	now := time.Unix(1000, 0)

	testCases := map[string]struct {
		addrConns         []addressConns
		addrIndex         int
		makeMetrics       func(ctrl *gomock.Controller) *MockMetrics
		expectedAddrConns []addressConns
	}{
		"no_cleanup_single_conn_live": {
			addrConns: []addressConns{{
				conns: []poolConn{
					{lastUsed: now},
				},
			}},
			expectedAddrConns: []addressConns{{
				conns: []poolConn{
					{lastUsed: now},
				},
			}},
		},
		"no_cleanup_single_conn_dead": {
			addrConns: []addressConns{{
				conns: []poolConn{
					{dead: true},
				},
			}},
			expectedAddrConns: []addressConns{{
				conns: []poolConn{
					{dead: true},
				},
			}},
		},
		"cleanup": {
			addrConns: []addressConns{{
				address: "127.0.0.1:853",
				conns: []poolConn{
					{lastUsed: now.Add(-maxIdleDuration - 1)}, // expired
					{inUse: true},
					{dead: true},
					{inUse: true},
					{dead: true},
					{lastUsed: now.Add(-maxIdleDuration - 2)}, // expired
					{lastUsed: now},
					{dead: true},
				},
			}},
			expectedAddrConns: []addressConns{{
				address: "127.0.0.1:853",
				conns: []poolConn{
					{lastUsed: now},
					{inUse: true},
					{dead: true},
					{inUse: true},
				},
			}},
			makeMetrics: func(ctrl *gomock.Controller) *MockMetrics {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().DeadConnsInc("127.0.0.1:853").Times(2)
				metrics.EXPECT().RemovedConnsAdd("127.0.0.1:853", uint(4))
				return metrics
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			setFieldsForAddrConns(testCase.addrConns)
			setFieldsForAddrConns(testCase.expectedAddrConns)

			pool := &Pool{
				addrConns: testCase.addrConns,
				timeNow:   func() time.Time { return now },
			}

			if testCase.makeMetrics != nil {
				ctrl := gomock.NewController(t)
				pool.metrics = testCase.makeMetrics(ctrl)
			}

			pool.cleanup(testCase.addrIndex)
			assert.Equal(t, testCase.expectedAddrConns, pool.addrConns)
		})
	}
}

// Fuzz_Pool_compact performs fuzz testing on the Pool.compact method,
// to make sure it always produces a sound cleaned up connections slice.
// Command: go test -fuzz=Fuzz_Pool_compact ./internal/pool
// Note it uses the [validateCompaction] to validate the result.
func Fuzz_Pool_compact(f *testing.F) {
	f.Fuzz(func(t *testing.T, length uint, inUse, dead uint8) {
		conns := make([]poolConn, length)
		for i := range conns {
			conns[i] = poolConn{
				connIndex: i,
				lastUsed:  time.Unix(int64(i), 0), // used as identifier
				inUse:     inUse%2 == 0,
			}
			if !conns[i].inUse {
				conns[i].dead = dead%2 == 0
			}
		}

		from := make([]poolConn, len(conns))
		copy(from, conns)

		pool := &Pool{
			timeNow: func() time.Time { return time.Unix(0, 0) }, // unused
		}

		to, removed := pool.compact(conns)

		validateCompaction(t, from, to, removed)
	})
}

// Note the lastUsed field is used as an identifier here.
// This function validates that:
// - in use connections are kept as is.
// - live not in use connections are kept, although they can be moved.
// - min(dead connections,connections-after-last-in-use) is removed from the end.
func validateCompaction(t *testing.T, from, to []poolConn, removed uint) {
	t.Helper()

	diff := uint(max(0, len(from)-len(to))) //nolint:gosec

	if removed != diff {
		t.Errorf("removed count %d does not match length difference %d",
			removed, diff)
	}

	toIDs := make(map[int64]struct{}, len(to))
	for _, conn := range to {
		toIDs[conn.lastUsed.Unix()] = struct{}{}
	}
	inUseConns := 0
	unusedLiveConns := 0
	lastInUseIndex := -1
	for i, conn := range from {
		switch {
		case conn.inUse:
			inUseConns++
			lastInUseIndex = i
			switch {
			case i >= len(to):
				t.Errorf("connection in use at index %d is out of bound for "+
					"compacted slice of length %d. From slice is:\n%s",
					i, len(to), connsToString(from))
			case to[i] != from[i]:
				t.Errorf("connection in use at index %d is not kept the same. "+
					" From slice is:\n%s", i, connsToString(from))
			}
		case !conn.dead:
			unusedLiveConns++
			fromID := conn.lastUsed.Unix()
			if _, found := toIDs[fromID]; !found {
				// Check live not in use connections are kept, although they can be moved.
				t.Errorf("live unused connection with id %d is not present in compacted slice "+
					" From slice is:\n%s", fromID, connsToString(from))
			}
		}
	}

	deadConns := len(from) - inUseConns - unusedLiveConns
	connsAfterLastInUse := len(from) - lastInUseIndex - 1
	expectedRemoved := min(deadConns, connsAfterLastInUse)

	if len(from)-len(to) != expectedRemoved {
		t.Errorf("wrong number of connections removed: got %d, want %d. From slice is:\n%s",
			len(from)-len(to), expectedRemoved, connsToString(from))
	}
}

func connsToString(conns []poolConn) string {
	s := "{\n"
	for _, conn := range conns {
		s += fmt.Sprintf("  {dead:%t, inUse:%t},\n", conn.dead, conn.inUse)
	}
	s += "\n}"
	return s
}
