package pool

import (
	"net"
	"time"
)

type poolConn struct {
	net.Conn
	addrIndex int
	connIndex int
	lastUsed  time.Time
	inUse     bool
	// dead is only used for caching for performance.
	dead bool
}

func (p *Pool) isConnDead(conn poolConn) bool {
	switch {
	case conn.dead:
		return true
	case conn.inUse:
		return false
	case p.timeNow().Sub(conn.lastUsed) <= p.maxIdleDuration:
		return false
	}
	conn.dead = true
	_ = conn.Close() // ignore error since it may already be closed.
	p.addrConns[conn.addrIndex].conns[conn.connIndex] = conn
	address := p.addrConns[conn.addrIndex].address
	p.metrics.DeadConnsInc(address)
	return true
}

func (p *Pool) addressFromConn(conn poolConn) string {
	return p.addrConns[conn.addrIndex].address
}
