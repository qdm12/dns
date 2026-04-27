package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_New(t *testing.T) {
	t.Parallel()

	t.Run("no_address", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		dialer := NewMockDialer(ctrl)
		dialer.EXPECT().Addresses().Return([]string{})
		assert.PanicsWithValue(t, "cannot create pool with zero addresses",
			func() { New(dialer, nil) })
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		testDialer := &testDialer{port: 8053}
		ctrl := gomock.NewController(t)
		testMetrics := NewMockMetrics(ctrl)

		expectedPool := &Pool{
			dialer:          testDialer,
			metrics:         testMetrics,
			addrConns:       []addressConns{{address: "127.0.0.1:8053", conns: []poolConn{}, connIDToIndex: map[uint64]int{}}},
			nextConnID:      1,
			maxIdleDuration: maxIdleDuration,
		}

		pool := New(testDialer, testMetrics)

		assert.NotNil(t, pool.timeNow)
		pool.timeNow = nil

		assert.Equal(t, expectedPool, pool)
	})
}

func Test_Pool_setIfAllAddrsHaveOneConn(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		pool         *Pool
		expectedPool *Pool
	}{
		"all_addresses_have_already_one_connection": {
			pool: &Pool{
				oneConnPerAddr: true,
			},
			expectedPool: &Pool{
				oneConnPerAddr: true,
			},
		},
		"all_addresses_but_one_have_one_connection": {
			pool: &Pool{
				addrConns: []addressConns{
					{conns: []poolConn{{}}},
					{conns: []poolConn{{}}},
					{conns: []poolConn{}},
				},
			},
			expectedPool: &Pool{
				addrConns: []addressConns{
					{conns: []poolConn{{}}},
					{conns: []poolConn{{}}},
					{conns: []poolConn{}},
				},
			},
		},
		"all_addresses_have_one_connection": {
			pool: &Pool{
				addrConns: []addressConns{
					{conns: []poolConn{{}}},
					{conns: []poolConn{{}}},
					{conns: []poolConn{{}}},
				},
			},
			expectedPool: &Pool{
				addrConns: []addressConns{
					{conns: []poolConn{{}}},
					{conns: []poolConn{{}}},
					{conns: []poolConn{{}}},
				},
				oneConnPerAddr: true,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pool := testCase.pool
			pool.setIfAllAddrsHaveOneConn()
			assert.Equal(t, testCase.expectedPool.oneConnPerAddr,
				pool.oneConnPerAddr)
		})
	}
}
