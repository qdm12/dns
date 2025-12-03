package pool

import (
	"context"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Pool_Renew(t *testing.T) {
	t.Parallel()

	t.Run("not_a_poolConn", func(t *testing.T) {
		t.Parallel()
		pool := &Pool{}
		conn := &net.TCPConn{}
		assert.PanicsWithValue(t, "cannot renew non-pool connection *net.TCPConn", func() {
			_, _ = pool.Renew(context.Background(), "tcp", conn)
		})
	})

	t.Run("connection_refused", func(t *testing.T) {
		t.Parallel()

		dialer := &testDialer{port: 0}
		ctrl := gomock.NewController(t)
		metrics := NewMockMetrics(ctrl)
		metrics.EXPECT().RenewRequestsInc("127.0.0.1:0")
		metrics.EXPECT().LiveConnsAdd("127.0.0.1:0", -1)
		metrics.EXPECT().RenewalsInc("127.0.0.1:0")
		pool := New(dialer, metrics)
		const network = "tcp"
		conn := (net.Conn)(poolConn{
			Conn:      &net.TCPConn{},
			addrIndex: 0,
			connIndex: 0,
		})
		conn, err := pool.Renew(context.Background(), network, conn)
		require.Error(t, err)
		assert.Equal(t, "dial tcp 127.0.0.1:0: connect: connection refused", err.Error())
		assert.Nil(t, conn)
	})

	t.Run("renew_dead_connection", func(t *testing.T) {
		t.Parallel()

		dialer, runErr := startLocalTCPServer(t)
		ctrl := gomock.NewController(t)
		metrics := NewMockMetrics(ctrl)
		addresses := dialer.Addresses()
		// Renew call metrics
		metrics.EXPECT().RenewRequestsInc(addresses[0])
		metrics.EXPECT().LiveConnsAdd(addresses[0], -1)
		// renew call metrics
		metrics.EXPECT().RenewalsInc(addresses[0])
		metrics.EXPECT().LiveConnsAdd(addresses[0], 1)
		deadConn := poolConn{
			Conn: &noopConn{},
			dead: true,
		}
		pool := &Pool{
			dialer:         dialer,
			metrics:        metrics,
			oneConnPerAddr: true,
			addrConns: []addressConns{
				{address: addresses[0], conns: []poolConn{deadConn}},
			},
		}
		ctx := context.Background()
		const network = "tcp"

		renewedConn, err := pool.Renew(ctx, network, deadConn)
		require.NoError(t, err)
		checkConnWorks(t, renewedConn)
		renewedPoolConn := renewedConn.(poolConn) //nolint:forcetypeassert
		renewedPoolConn.Conn = nil                // remove Conn from comparison
		expectedPoolConn := poolConn{}
		assert.Equal(t, expectedPoolConn, renewedPoolConn)

		select {
		case err := <-runErr:
			require.NoError(t, err)
		default:
		}
	})

	t.Run("renew_live_connection", func(t *testing.T) {
		t.Parallel()
		// this should not happen in practice, but test it anyway

		dialer, runErr := startLocalTCPServer(t)
		ctrl := gomock.NewController(t)
		metrics := NewMockMetrics(ctrl)
		addresses := dialer.Addresses()
		// Renew call metrics
		metrics.EXPECT().RenewRequestsInc(addresses[0])
		metrics.EXPECT().LiveConnsAdd(addresses[0], -1)
		// renew call metrics
		metrics.EXPECT().RenewalsInc(addresses[0])
		metrics.EXPECT().LiveConnsAdd(addresses[0], 1)
		netConn, err := dialer.Dial(context.Background(), "tcp", addresses[0])
		require.NoError(t, err)
		liveConn := poolConn{
			Conn: netConn,
			dead: true,
		}
		pool := &Pool{
			dialer:         dialer,
			metrics:        metrics,
			oneConnPerAddr: true,
			addrConns: []addressConns{
				{address: addresses[0], conns: []poolConn{liveConn}},
			},
		}
		ctx := context.Background()
		const network = "tcp"

		renewedConn, err := pool.Renew(ctx, network, liveConn)
		require.NoError(t, err)
		checkConnWorks(t, renewedConn)
		renewedPoolConn := renewedConn.(poolConn) //nolint:forcetypeassert
		renewedPoolConn.Conn = nil                // remove Conn from comparison
		expectedPoolConn := poolConn{}
		assert.Equal(t, expectedPoolConn, renewedPoolConn)

		select {
		case err := <-runErr:
			require.NoError(t, err)
		default:
		}
	})
}
