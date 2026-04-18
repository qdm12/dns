package pool

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Pool_Renew(t *testing.T) {
	t.Parallel()

	now := time.Unix(10000, 0)
	timeNow := func() time.Time {
		return now
	}

	t.Run("not_a_poolConn", func(t *testing.T) {
		t.Parallel()
		pool := &Pool{}
		conn := &noopConn{}
		assert.PanicsWithValue(t, "cannot renew non-pool connection *pool.noopConn", func() {
			_, _ = pool.Renew(context.Background(), "tcp", conn)
		})
	})

	t.Run("connection_refused", func(t *testing.T) {
		t.Parallel()

		dialer := &testDialer{port: 0}
		ctrl := gomock.NewController(t)
		metrics := NewMockMetrics(ctrl)
		address := dialer.Addresses()[0]
		// [Pool.Renew] metric calls
		metrics.EXPECT().DeadConnInc(address)
		// [Pool.renew] metric calls
		metrics.EXPECT().NewConnsInc(address, outcomeError)
		metrics.EXPECT().RenewConnInc(address, renewReasonConnError, outcomeError)
		metrics.EXPECT().RecordLifetime(address, time.Minute)
		pool := &Pool{
			dialer:         dialer,
			metrics:        metrics,
			timeNow:        timeNow,
			oneConnPerAddr: true,
			addrConns: []addressConns{{
				address: address,
				conns:   []poolConn{{}},
			}},
		}
		expectedPool := &Pool{
			dialer:         dialer,
			oneConnPerAddr: true,
			addrConns: []addressConns{{
				address: address,
				conns: []poolConn{{
					Conn:    &noopConn{},
					created: now.Add(-time.Minute),
					dead:    true,
				}},
			}},
		}

		const network = "tcp"
		netConn := (net.Conn)(poolConn{
			Conn:    &noopConn{},
			created: now.Add(-time.Minute),
		})
		netConn, err := pool.Renew(context.Background(), network, netConn)
		require.Error(t, err)
		assert.Equal(t, "dial tcp 127.0.0.1:0: connect: connection refused", err.Error())
		assert.Nil(t, netConn)
		clearPoolFieldsForComparison(pool)
		assert.Equal(t, expectedPool, pool)
	})

	t.Run("renew_dead_connection", func(t *testing.T) {
		t.Parallel()

		dialer, runErr := startLocalTCPServer(t, handleConnCopy)
		address := dialer.Addresses()[0]
		ctrl := gomock.NewController(t)
		metrics := NewMockMetrics(ctrl)
		// [Pool.renew] metric calls
		metrics.EXPECT().NewConnsInc(address, outcomeSuccess)
		metrics.EXPECT().RenewConnInc(address, renewReasonConnError, outcomeSuccess)
		metrics.EXPECT().RecordLifetime(address, time.Minute)
		deadConn := poolConn{
			Conn:    &noopConn{},
			created: now.Add(-time.Minute),
			dead:    true,
		}
		pool := &Pool{
			dialer:         dialer,
			metrics:        metrics,
			timeNow:        timeNow,
			oneConnPerAddr: true,
			addrConns: []addressConns{{
				address: address,
				conns:   []poolConn{deadConn},
			}},
		}
		expectedPool := &Pool{
			dialer:         dialer,
			oneConnPerAddr: true,
			addrConns: []addressConns{{
				address: address,
				conns:   []poolConn{{created: now, lastUsed: now, inUse: true}},
			}},
		}
		ctx := context.Background()
		const network = "tcp"

		renewedConn, err := pool.Renew(ctx, network, deadConn)
		require.NoError(t, err)
		checkConnCopies(t, renewedConn)
		renewedPoolConn := renewedConn.(poolConn) //nolint:forcetypeassert
		// remove Conn from comparison
		renewedPoolConn.Conn = nil
		pool.addrConns[renewedPoolConn.addrIndex].conns[renewedPoolConn.connIndex].Conn = nil
		expectedPoolConn := poolConn{
			created:  now,
			lastUsed: now,
			inUse:    true,
		}
		assert.Equal(t, expectedPoolConn, renewedPoolConn)
		clearPoolFieldsForComparison(pool)
		assert.Equal(t, expectedPool, pool)

		select {
		case err := <-runErr:
			require.NoError(t, err)
		default:
		}
	})

	t.Run("renew_live_connection", func(t *testing.T) {
		t.Parallel()
		// this should not happen in practice, but test it anyway

		dialer, runErr := startLocalTCPServer(t, handleConnCopy)
		ctrl := gomock.NewController(t)
		metrics := NewMockMetrics(ctrl)
		address := dialer.Addresses()[0]
		// [Pool.renew] metric calls
		metrics.EXPECT().NewConnsInc(address, outcomeSuccess)
		metrics.EXPECT().RenewConnInc(address, renewReasonConnError, outcomeSuccess)
		metrics.EXPECT().RecordLifetime(address, time.Minute)
		netConn, err := dialer.Dial(context.Background(), "tcp", address)
		require.NoError(t, err)
		liveConn := poolConn{
			Conn:    netConn,
			created: now.Add(-time.Minute),
		}
		pool := &Pool{
			dialer:         dialer,
			metrics:        metrics,
			timeNow:        timeNow,
			oneConnPerAddr: true,
			addrConns: []addressConns{
				{address: address, conns: []poolConn{liveConn}},
			},
		}
		ctx := context.Background()
		const network = "tcp"

		renewedConn, err := pool.Renew(ctx, network, liveConn)
		require.NoError(t, err)
		checkConnCopies(t, renewedConn)
		renewedPoolConn := renewedConn.(poolConn) //nolint:forcetypeassert
		renewedPoolConn.Conn = nil                // remove Conn from comparison
		expectedPoolConn := poolConn{
			created:  now,
			lastUsed: now,
			inUse:    true,
		}
		assert.Equal(t, expectedPoolConn, renewedPoolConn)

		select {
		case err := <-runErr:
			require.NoError(t, err)
		default:
		}
	})
}
