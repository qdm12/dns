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

func Test_Pool_Get(t *testing.T) { //nolint:maintidx
	t.Parallel()

	now := time.Unix(10000, 0)
	dialer, runErr := startLocalTCPServer(t)

	testCases := map[string]struct {
		makePool     func(ctrl *gomock.Controller) *Pool
		conn         net.Conn
		errMessage   string
		expectedPool *Pool
	}{
		"error_on_first_connection_of_address": {
			makePool: func(*gomock.Controller) *Pool {
				return &Pool{
					dialer: dialer,
					addrConns: []addressConns{
						{address: "127.0.0.1:0", conns: []poolConn{}},
					},
				}
			},
			errMessage: "creating connection: dial tcp 127.0.0.1:0: connect: connection refused",
			expectedPool: &Pool{
				dialer: dialer,
				addrConns: []addressConns{
					{address: "127.0.0.1:0", conns: []poolConn{}},
				},
			},
		},
		"new_connection_for_address_without_connection": {
			makePool: func(ctrl *gomock.Controller) *Pool {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().ConnsAdd(dialer.Addresses()[0], 1)
				metrics.EXPECT().LiveConnsAdd(dialer.Addresses()[0], 1)
				metrics.EXPECT().InUseConnsAdd(dialer.Addresses()[0], 1)
				return &Pool{
					dialer:  dialer,
					metrics: metrics,
					addrConns: []addressConns{
						{address: "already_has_a_conn", conns: []poolConn{{}}},
						{address: "already_has_a_conn", conns: []poolConn{{}}}, // start here
						{address: dialer.Addresses()[0], conns: []poolConn{}},
					},
				}
			},
			conn: poolConn{
				addrIndex: 2,
				connIndex: 0,
				inUse:     true,
			},
			expectedPool: &Pool{
				dialer:            dialer,
				lastUsedAddrIndex: 2,
				oneConnPerAddr:    true,
				addrConns: []addressConns{
					{address: "already_has_a_conn", conns: []poolConn{{}}},
					{address: "already_has_a_conn", conns: []poolConn{{}}},
					{address: dialer.Addresses()[0], conns: []poolConn{{
						addrIndex: 2,
						connIndex: 0,
						inUse:     true,
					}}},
				},
			},
		},
		"all_connections_are_in_use": {
			makePool: func(ctrl *gomock.Controller) *Pool {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().ConnsAdd(dialer.Addresses()[0], 1)
				metrics.EXPECT().LiveConnsAdd(dialer.Addresses()[0], 1)
				metrics.EXPECT().InUseConnsAdd(dialer.Addresses()[0], 1)
				return &Pool{
					dialer:         dialer,
					metrics:        metrics,
					oneConnPerAddr: true,
					addrConns: []addressConns{
						{address: "previously_used_address", conns: []poolConn{{inUse: true}}},
						{address: "two_connections_in_use", conns: []poolConn{{inUse: true}, {inUse: true}}}, // start here
						// Pick this one because it's the closest with minimum connections
						{address: dialer.Addresses()[0], conns: []poolConn{{inUse: true}}},
						{address: "too_far_from_next", conns: []poolConn{{inUse: true}}},
					},
				}
			},
			conn: poolConn{
				addrIndex: 2,
				connIndex: 1,
				inUse:     true,
			},
			expectedPool: &Pool{
				dialer:            dialer,
				lastUsedAddrIndex: 2,
				oneConnPerAddr:    true,
				addrConns: []addressConns{
					{address: "previously_used_address", conns: []poolConn{{inUse: true}}},
					{address: "two_connections_in_use", conns: []poolConn{{inUse: true}, {inUse: true}}},
					{address: dialer.Addresses()[0], conns: []poolConn{{inUse: true}, {
						addrIndex: 2,
						connIndex: 1,
						inUse:     true,
					}}},
					{address: "too_far_from_next", conns: []poolConn{{inUse: true}}},
				},
			},
		},
		"error_on_new_and_all_connections_are_in_use": {
			makePool: func(*gomock.Controller) *Pool {
				return &Pool{
					dialer:         dialer,
					oneConnPerAddr: true,
					addrConns: []addressConns{
						{address: "previously_used_address", conns: []poolConn{{inUse: true}}},
						{address: "127.0.0.1:0", conns: []poolConn{{inUse: true}}},
					},
				}
			},
			errMessage: "creating connection: dial tcp 127.0.0.1:0: connect: connection refused",
			expectedPool: &Pool{
				dialer:         dialer,
				oneConnPerAddr: true,
				addrConns: []addressConns{
					{address: "previously_used_address", conns: []poolConn{{inUse: true}}},
					{address: "127.0.0.1:0", conns: []poolConn{{inUse: true}}},
				},
			},
		},
		"found_live_connection": {
			makePool: func(ctrl *gomock.Controller) *Pool {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().InUseConnsAdd(dialer.Addresses()[0], 1)

				liveNetConn, err := dialer.Dial(context.Background(), "tcp", dialer.Addresses()[0])
				require.NoError(t, err)
				liveConn := poolConn{
					Conn:      liveNetConn,
					addrIndex: 1,
					connIndex: 0,
					inUse:     false,
					lastUsed:  now,
				}

				return &Pool{
					dialer:         dialer,
					metrics:        metrics,
					oneConnPerAddr: true,
					timeNow:        func() time.Time { return now },
					addrConns: []addressConns{
						{address: "previously_used", conns: []poolConn{{}}},
						{address: dialer.Addresses()[0], conns: []poolConn{liveConn}}, // start here
					},
				}
			},
			conn: poolConn{
				addrIndex: 1,
				connIndex: 0,
				inUse:     true,
				lastUsed:  now,
			},
			expectedPool: &Pool{
				dialer:            dialer,
				lastUsedAddrIndex: 1,
				oneConnPerAddr:    true,
				addrConns: []addressConns{
					{address: "previously_used", conns: []poolConn{{}}},
					{address: dialer.Addresses()[0], conns: []poolConn{{
						addrIndex: 1,
						connIndex: 0,
						inUse:     true,
						lastUsed:  now,
					}}},
				},
			},
		},
		"found_dead_connection_renew_error": {
			makePool: func(ctrl *gomock.Controller) *Pool {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().RenewalsInc("127.0.0.1:0")

				return &Pool{
					dialer:         dialer,
					metrics:        metrics,
					oneConnPerAddr: true,
					addrConns: []addressConns{
						{address: "previously_used", conns: []poolConn{{}}},
						{address: "127.0.0.1:0", conns: []poolConn{{
							Conn:      &noopConn{},
							addrIndex: 1,
							connIndex: 0,
							dead:      true,
						}}}, // start here
					},
				}
			},
			errMessage: "renewing dead connection: dial tcp 127.0.0.1:0: connect: connection refused",
			expectedPool: &Pool{
				dialer:            dialer,
				oneConnPerAddr:    true,
				lastUsedAddrIndex: 1,
				addrConns: []addressConns{
					{address: "previously_used", conns: []poolConn{{}}},
					{address: "127.0.0.1:0", conns: []poolConn{{
						Conn:      &noopConn{},
						addrIndex: 1,
						connIndex: 0,
						dead:      true,
					}}},
				},
			},
		},
		"found_dead_connection_renew_success": {
			makePool: func(ctrl *gomock.Controller) *Pool {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().RenewalsInc(dialer.Addresses()[0])
				metrics.EXPECT().LiveConnsAdd(dialer.Addresses()[0], 1)
				metrics.EXPECT().InUseConnsAdd(dialer.Addresses()[0], 1)

				return &Pool{
					dialer:         dialer,
					metrics:        metrics,
					oneConnPerAddr: true,
					timeNow:        func() time.Time { return now },
					addrConns: []addressConns{
						{address: "previously_used", conns: []poolConn{{}}},
						{address: dialer.Addresses()[0], conns: []poolConn{{
							Conn:      &noopConn{},
							addrIndex: 1,
							connIndex: 0,
							dead:      true,
						}}}, // start here
					},
				}
			},
			conn: poolConn{
				addrIndex: 1,
				connIndex: 0,
				inUse:     true,
			},
			expectedPool: &Pool{
				dialer:            dialer,
				lastUsedAddrIndex: 1,
				oneConnPerAddr:    true,
				addrConns: []addressConns{
					{address: "previously_used", conns: []poolConn{{}}},
					{address: dialer.Addresses()[0], conns: []poolConn{{
						addrIndex: 1,
						connIndex: 0,
						inUse:     true,
					}}},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			pool := testCase.makePool(ctrl)

			ctx := context.Background()
			const network = "tcp"
			conn, err := pool.Get(ctx, network)

			if testCase.errMessage != "" {
				assert.EqualError(t, err, testCase.errMessage)
			} else {
				require.NoError(t, err)
				checkConnWorks(t, conn)
				poolConn := conn.(poolConn) //nolint:forcetypeassert
				poolConn.Conn = nil         // ignore net.Conn comparison
				conn = poolConn
				pool.addrConns[poolConn.addrIndex].conns[poolConn.connIndex].Conn = nil
			}
			assert.Equal(t, testCase.conn, conn)
			pool.metrics = nil // ignore metrics comparison
			pool.timeNow = nil // ignore timeNow comparison
			assert.Equal(t, testCase.expectedPool, pool)
		})
	}
	select {
	case err := <-runErr:
		require.NoError(t, err)
	default:
	}
}

func Test_Pool_findNextAvailConn(t *testing.T) {
	t.Parallel()

	now := time.Unix(10000, 0)

	testCases := map[string]struct {
		addrConns         []addressConns
		lastUsedAddrIndex int
		conn              poolConn
		found             bool
		live              bool
	}{
		"single_address_no_conn": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{}},
			},
			found: false,
		},
		"single_address_single_conn": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{{lastUsed: now}}},
			},
			conn:  poolConn{lastUsed: now},
			found: true,
			live:  true,
		},
		"single_address_single_conn_in_use": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{{inUse: true}}},
			},
			found: false,
		},
		"single_address_single_conn_dead": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{{dead: true}}},
			},
			conn:  poolConn{dead: true},
			found: true,
			live:  false,
		},
		"single_address_multiple_conns_get_live": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{
					{inUse: true},
					{dead: true},
					{lastUsed: now}, // live available
					{dead: true},
				}},
			},
			conn:  poolConn{lastUsed: now},
			found: true,
			live:  true,
		},
		"single_address_multiple_conns_get_dead": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{
					{inUse: true},
					{dead: true, lastUsed: now}, // use the first dead found if no live available
					{dead: true},
				}},
			},
			conn:  poolConn{dead: true, lastUsed: now},
			found: true,
			live:  false,
		},
		"multiple_addresses_use_from_last_used": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{{lastUsed: now}}}, // previously used, but only available
				{address: "B", conns: []poolConn{{inUse: true}}},   // checked first, all in use.
				{address: "C", conns: []poolConn{{inUse: true}}},   // checked second, all in use.
			},
			lastUsedAddrIndex: 0,
			conn:              poolConn{lastUsed: now},
			found:             true,
			live:              true,
		},
		"multiple_addresses_use_next_live": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{{lastUsed: now.Add(-time.Second)}}}, // previously used
				{address: "B", conns: []poolConn{{inUse: true}}},                     // checked first, all in use.
				{address: "C", conns: []poolConn{{lastUsed: now}}},                   // checked second, use this one.
			},
			lastUsedAddrIndex: 0,
			conn:              poolConn{lastUsed: now},
			found:             true,
			live:              true,
		},
		"multiple_addresses_use_next_dead": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{{lastUsed: now.Add(-time.Second)}}}, // previously used
				{address: "B", conns: []poolConn{{inUse: true}}},                     // checked first, all in use.
				{address: "C", conns: []poolConn{{dead: true}}},                      // checked second, use this one.
			},
			lastUsedAddrIndex: 0,
			conn:              poolConn{dead: true},
			found:             true,
			live:              false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Less(t, testCase.lastUsedAddrIndex, len(testCase.addrConns),
				"lastUsedAddrIndex must be less than number of addresses")

			pool := &Pool{
				addrConns:         testCase.addrConns,
				lastUsedAddrIndex: testCase.lastUsedAddrIndex,
				timeNow: func() time.Time {
					return now
				},
				maxIdleDuration: maxIdleDuration,
			}

			conn, found, live := pool.findNextAvailConn()

			assert.Equal(t, testCase.conn, conn)
			assert.Equal(t, testCase.found, found)
			assert.Equal(t, testCase.live, live)
		})
	}
}

func Test_Pool_newConn(t *testing.T) {
	t.Parallel()

	dialer, runErr := startLocalTCPServer(t)

	testCases := map[string]struct {
		makePool     func(ctrl *gomock.Controller) *Pool
		conn         poolConn
		errMessage   string
		expectedPool *Pool
	}{
		"dial_error": {
			makePool: func(*gomock.Controller) *Pool {
				return &Pool{
					dialer: dialer,
					addrConns: []addressConns{
						{address: "127.0.0.1:0"},
					},
				}
			},
			errMessage: "dial tcp 127.0.0.1:0: connect: connection refused",
			expectedPool: &Pool{
				dialer: dialer,
				addrConns: []addressConns{
					{address: "127.0.0.1:0"},
				},
			},
		},
		"new_conn_for_next_address_without_conn": {
			makePool: func(ctrl *gomock.Controller) *Pool {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().ConnsAdd(dialer.Addresses()[0], 1)
				metrics.EXPECT().LiveConnsAdd(dialer.Addresses()[0], 1)
				return &Pool{
					dialer:  dialer,
					metrics: metrics,
					addrConns: []addressConns{
						{address: "previously_used_address", conns: []poolConn{{}}},
						{address: "already_have_one_conn", conns: []poolConn{{}}}, // start here
						{address: dialer.Addresses()[0], conns: []poolConn{}},     // no conns, use this one
						{address: "no_conns_either", conns: []poolConn{}},
					},
				}
			},
			conn: poolConn{
				addrIndex: 2,
				connIndex: 0,
				inUse:     true,
			},
			expectedPool: &Pool{
				dialer:            dialer,
				lastUsedAddrIndex: 2,
				addrConns: []addressConns{
					{address: "previously_used_address", conns: []poolConn{{}}},
					{address: "already_have_one_conn", conns: []poolConn{{}}},
					{address: dialer.Addresses()[0], conns: []poolConn{{
						addrIndex: 2,
						connIndex: 0,
						inUse:     true,
					}}},
					{address: "no_conns_either", conns: []poolConn{}},
				},
			},
		},
		"new_conn_for_current_address": {
			makePool: func(ctrl *gomock.Controller) *Pool {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().ConnsAdd(dialer.Addresses()[0], 1)
				metrics.EXPECT().LiveConnsAdd(dialer.Addresses()[0], 1)
				return &Pool{
					dialer:         dialer,
					metrics:        metrics,
					oneConnPerAddr: true,
					addrConns: []addressConns{
						{address: "previously_used_address", conns: []poolConn{{}}},
						{address: dialer.Addresses()[0], conns: []poolConn{{}}}, // start here
					},
				}
			},
			conn: poolConn{
				addrIndex: 1,
				connIndex: 1,
				inUse:     true,
			},
			expectedPool: &Pool{
				dialer:            dialer,
				lastUsedAddrIndex: 1,
				oneConnPerAddr:    true,
				addrConns: []addressConns{
					{address: "previously_used_address", conns: []poolConn{{}}},
					{address: dialer.Addresses()[0], conns: []poolConn{{}, {
						addrIndex: 1,
						connIndex: 1,
						inUse:     true,
					}}},
				},
			},
		},
		"new_conn_for_next_address": {
			makePool: func(ctrl *gomock.Controller) *Pool {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().ConnsAdd(dialer.Addresses()[0], 1)
				metrics.EXPECT().LiveConnsAdd(dialer.Addresses()[0], 1)
				return &Pool{
					dialer:         dialer,
					metrics:        metrics,
					oneConnPerAddr: true,
					addrConns: []addressConns{
						{address: "previously_used_address", conns: []poolConn{{}}},
						{address: "not_min_conns", conns: []poolConn{{}, {}}},   // start here
						{address: dialer.Addresses()[0], conns: []poolConn{{}}}, // pick that one
						{address: "not_min_conns", conns: []poolConn{{}, {}}},
					},
				}
			},
			conn: poolConn{
				addrIndex: 2,
				connIndex: 1,
				inUse:     true,
			},
			expectedPool: &Pool{
				dialer:            dialer,
				lastUsedAddrIndex: 2,
				oneConnPerAddr:    true,
				addrConns: []addressConns{
					{address: "previously_used_address", conns: []poolConn{{}}},
					{address: "not_min_conns", conns: []poolConn{{}, {}}}, // start here
					{address: dialer.Addresses()[0], conns: []poolConn{{}, {
						addrIndex: 2,
						connIndex: 1,
						inUse:     true,
					}}},
					{address: "not_min_conns", conns: []poolConn{{}, {}}},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			pool := testCase.makePool(ctrl)

			ctx := context.Background()
			const network = "tcp"
			conn, err := pool.newConn(ctx, network)

			if testCase.errMessage != "" {
				assert.EqualError(t, err, testCase.errMessage)
			} else {
				require.NoError(t, err)
				checkConnWorks(t, conn.Conn)
				conn.Conn = nil // ignore net.Conn comparison
				pool.addrConns[conn.addrIndex].conns[conn.connIndex].Conn = nil
			}
			assert.Equal(t, testCase.conn, conn)
			pool.metrics = nil // ignore metrics comparison
			assert.Equal(t, testCase.expectedPool, pool)
		})
	}
	select {
	case err := <-runErr:
		require.NoError(t, err)
	default:
	}
}

func Test_Pool_findAddressForNewConn(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		addrConns         []addressConns
		lastUsedAddrIndex int
		addressIndex      int
	}{
		"single_address_no_conns": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{}},
			},
			lastUsedAddrIndex: 0,
			addressIndex:      0,
		},
		"single_address_with_conns": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{{}}},
			},
			lastUsedAddrIndex: 0,
			addressIndex:      0,
		},
		"addresses_with_no_conn_last_index_0": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{}},
				{address: "B", conns: []poolConn{}},
				{address: "C", conns: []poolConn{}},
			},
			lastUsedAddrIndex: 0,
			addressIndex:      1,
		},
		"addresses_with_no_conn_last_index_last": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{}},
				{address: "B", conns: []poolConn{}},
				{address: "C", conns: []poolConn{}},
			},
			lastUsedAddrIndex: 2,
			addressIndex:      0,
		},
		"addresses_with_one_conn_last_index_0": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{{}}},
				{address: "B", conns: []poolConn{{}}},
				{address: "C", conns: []poolConn{{}}},
			},
			lastUsedAddrIndex: 0,
			addressIndex:      1,
		},
		"addresses_with_different_conns_last_index_0": {
			addrConns: []addressConns{
				{address: "A", conns: []poolConn{{}}},
				{address: "B", conns: []poolConn{{}, {}}},
				{address: "C", conns: []poolConn{{}, {}, {}}},
				{address: "D", conns: []poolConn{{}}},
				{address: "E", conns: []poolConn{{}, {}}},
			},
			lastUsedAddrIndex: 0,
			addressIndex:      3,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Less(t, testCase.lastUsedAddrIndex, len(testCase.addrConns),
				"lastUsedAddrIndex must be less than number of addresses")

			pool := &Pool{
				addrConns:         testCase.addrConns,
				lastUsedAddrIndex: testCase.lastUsedAddrIndex,
			}

			addressIndex := pool.findAddressForNewConn()

			assert.Equal(t, testCase.addressIndex, addressIndex)
		})
	}
}
