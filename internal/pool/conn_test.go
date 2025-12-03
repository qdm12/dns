package pool

import (
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_Pool_isConnDead(t *testing.T) {
	t.Parallel()

	now := time.Unix(1000, 0)

	testCases := map[string]struct {
		pool         *Pool
		expectedPool *Pool
		conn         poolConn
		dead         bool
		makeMetrics  func(ctrl *gomock.Controller) *MockMetrics
	}{
		"connection_already_dead": {
			pool:         &Pool{},
			expectedPool: &Pool{},
			conn:         poolConn{dead: true},
			dead:         true,
		},
		"connection_in_use": {
			pool:         &Pool{},
			expectedPool: &Pool{},
			conn:         poolConn{inUse: true},
		},
		"connection_not_expired": {
			pool: &Pool{
				maxIdleDuration: maxIdleDuration,
			},
			expectedPool: &Pool{
				maxIdleDuration: maxIdleDuration,
			},
			conn: poolConn{
				lastUsed: now.Add(-maxIdleDuration),
			},
		},
		"connection_expired": {
			pool: &Pool{
				maxIdleDuration: maxIdleDuration,
				addrConns: []addressConns{
					{},
					{
						address: "127.0.0.1:853",
						conns:   []poolConn{{}, {}},
					},
				},
			},
			expectedPool: &Pool{
				maxIdleDuration: maxIdleDuration,
				addrConns: []addressConns{{}, {
					address: "127.0.0.1:853",
					conns: []poolConn{{}, {
						Conn:      &net.TCPConn{},
						addrIndex: 1,
						connIndex: 1,
						dead:      true,
						lastUsed:  now.Add(-maxIdleDuration - time.Second),
					}},
				}},
			},
			conn: poolConn{
				Conn:      &net.TCPConn{},
				addrIndex: 1,
				connIndex: 1,
				lastUsed:  now.Add(-maxIdleDuration - time.Second),
			},
			dead: true,
			makeMetrics: func(ctrl *gomock.Controller) *MockMetrics {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().LiveConnsAdd("127.0.0.1:853", -1)
				return metrics
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pool := testCase.pool
			pool.timeNow = func() time.Time {
				return now
			}
			if testCase.makeMetrics != nil {
				ctrl := gomock.NewController(t)
				pool.metrics = testCase.makeMetrics(ctrl)
				testCase.expectedPool.metrics = pool.metrics
			}
			conn := testCase.conn

			dead := pool.isConnDead(conn)

			assert.Equal(t, testCase.dead, dead)
			pool.timeNow = nil
			assert.Equal(t, testCase.expectedPool, pool)
		})
	}
}
