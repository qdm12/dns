package pool

import (
	"net"
	"time"
)

type poolConn struct {
	net.Conn
	id        uint64
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
	p.setConn(conn)

	address := p.addrConns[conn.addrIndex].address
	p.metrics.DeadConnInc(address)
	lifetime := now.Sub(conn.created)
	p.metrics.RecordLifetime(address, lifetime)

	return true
}

func (p *Pool) addressFromConn(conn poolConn) string {
	return p.addrConns[conn.addrIndex].address
}

func (p *Pool) setConn(conn poolConn) {
	addrConns := &p.addrConns[conn.addrIndex]
	if conn.id == 0 {
		storedConn := addrConns.conns[conn.connIndex]
		if storedConn.id != 0 {
			conn.id = storedConn.id
		}
		addrConns.conns[conn.connIndex] = conn
		p.ensureConnIDToIndex(addrConns)
		addrConns.connIDToIndex[conn.id] = conn.connIndex
		return
	}
	p.ensureConnIDToIndex(addrConns)
	connIndex := addrConns.connIDToIndex[conn.id]
	conn.connIndex = connIndex
	addrConns.conns[connIndex] = conn
}

func (p *Pool) connFromID(addrIndex int, id uint64) (conn poolConn, found bool) {
	addrConns := &p.addrConns[addrIndex]
	if id == 0 {
		return poolConn{}, false
	}
	p.ensureConnIDToIndex(addrConns)
	connIndex, found := addrConns.connIDToIndex[id]
	if !found {
		return poolConn{}, false
	}
	conn = addrConns.conns[connIndex]
	conn.connIndex = connIndex
	return conn, true
}

func (p *Pool) ensureConnIDToIndex(addrConns *addressConns) {
	if addrConns.connIDToIndex != nil {
		return
	}
	addrConns.connIDToIndex = make(map[uint64]int, len(addrConns.conns))
	for i, conn := range addrConns.conns {
		_, duplicate := addrConns.connIDToIndex[conn.id]
		if conn.id == 0 || duplicate {
			conn.id = p.nextID()
			addrConns.conns[i] = conn
		}
		addrConns.connIDToIndex[conn.id] = i
	}
}

func (p *Pool) rebuildConnIDToIndex(addrIndex int) {
	addrConns := &p.addrConns[addrIndex]
	addrConns.connIDToIndex = make(map[uint64]int, len(addrConns.conns))
	for i, conn := range addrConns.conns {
		if conn.id == 0 {
			conn.id = p.nextID()
		}
		conn.connIndex = i
		addrConns.conns[i] = conn
		addrConns.connIDToIndex[conn.id] = i
	}
}

func (p *Pool) nextID() uint64 {
	if p.nextConnID == 0 {
		p.nextConnID = 1
	}
	id := p.nextConnID
	p.nextConnID++
	return id
}
