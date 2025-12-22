package pool

import (
	"net"
	"time"
)

type poolConn struct {
	net.Conn
	addrIndex int
	connIndex int
	// created is the time the connection was created.
	// It is only used for metrics.
	created time.Time
	// lastUsed is set to the current time when the connection is get
	// from the pool with [Pool.Get] and when it is put back with [Pool.Put].
	lastUsed time.Time
	inUse    bool
	// dead is only used for caching for performance.
	dead bool
}

// isConnDead returns whether the connection is dead.
// It updates the connection state in the pool if it is dead,
// as well as metrics.
func (p *Pool) isConnDead(conn poolConn) bool {
	switch {
	case conn.dead:
		return true
	case conn.inUse:
		return false
	}

	now := p.timeNow()
	if now.Sub(conn.lastUsed) <= p.maxIdleDuration {
		return false
	}

	conn.dead = true
	_ = conn.Close() // ignore error since it may already be closed.
	p.addrConns[conn.addrIndex].conns[conn.connIndex] = conn

	address := p.addrConns[conn.addrIndex].address
	p.metrics.DeadConnInc(address)
	lifetime := now.Sub(conn.created)
	p.metrics.RecordLifetime(address, lifetime)

	return true
}

func (p *Pool) addressFromConn(conn poolConn) string {
	return p.addrConns[conn.addrIndex].address
}
