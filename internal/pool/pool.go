package pool

import (
	"sync"
	"time"
)

type Pool struct {
	dialer  Dialer
	metrics Metrics

	// Internal state
	mutex             sync.Mutex
	addrConns         []addressConns
	lastUsedAddrIndex int
	oneConnPerAddr    bool

	timeNow         func() time.Time
	maxIdleDuration time.Duration
}

type addressConns struct {
	address string
	conns   []poolConn
}

const maxIdleDuration = 2 * time.Hour

const (
	outcomeSuccess = "success"
	outcomeError   = "error"
)

// New creates a new connection pool which uses the dialer
// to create new connections, and reports metrics using the
// metrics argument.
func New(dialer Dialer, metrics Metrics) *Pool {
	addresses := dialer.Addresses()
	if len(addresses) == 0 {
		panic("cannot create pool with zero addresses")
	}
	addrConns := make([]addressConns, len(addresses))
	for i, address := range addresses {
		addrConns[i] = addressConns{
			address: address,
			conns:   []poolConn{},
		}
	}
	return &Pool{
		dialer:          dialer,
		metrics:         metrics,
		addrConns:       addrConns,
		timeNow:         time.Now,
		maxIdleDuration: maxIdleDuration,
	}
}

func (p *Pool) setIfAllAddrsHaveOneConn() {
	for _, addrConns := range p.addrConns {
		if len(addrConns.conns) == 0 {
			p.oneConnPerAddr = false
			return
		}
	}
	p.oneConnPerAddr = true
}
