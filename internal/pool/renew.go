package pool

import (
	"context"
	"fmt"
	"net"
)

const (
	renewReasonMarkedDead = "marked dead"
	renewReasonConnError  = "connection error"
)

// Renew creates a new connection to replace the given one in the pool.
// It should be used when the caller detects that the given connection is dead,
// for example after receiving a io.EOF error.
// If Renew fails, the connection returned still has internal fields which must be used
// in a call to [Pool.PutDead] so the pool marks it as dead.
func (p *Pool) Renew(ctx context.Context, network string, conn net.Conn) (newConn net.Conn, err error) {
	poolConn, ok := conn.(poolConn)
	if !ok {
		panic(fmt.Sprintf("cannot renew non-pool connection %T", conn))
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	address := p.addressFromConn(poolConn)
	poolConn, err = p.renew(ctx, poolConn, network, renewReasonConnError)
	if err != nil {
		p.metrics.DeadConnInc(address)
		return nil, err
	}
	return poolConn, nil
}

// renew creates a new connection to replace the given one in the pool.
// It assumes the pool mutex is already held.
func (p *Pool) renew(ctx context.Context, conn poolConn, network, reason string) (
	newConn poolConn, err error,
) {
	_ = conn.Close() // ignore error since it may already be closed.
	address := p.addressFromConn(conn)
	defer func() {
		outcome := outcomeSuccess
		if err != nil {
			outcome = outcomeError
		}
		p.metrics.NewConnsInc(address, outcome)
		p.metrics.RenewConnInc(address, reason, outcome)
	}()
	netConn, err := p.dialer.Dial(ctx, network, address)
	if err != nil {
		// The pool will retry renewing the connection on another Get call.
		conn.inUse = false
		conn.dead = true
		p.addrConns[conn.addrIndex].conns[conn.connIndex] = conn
		return poolConn{}, err
	}
	conn.Conn = netConn
	conn.dead = false
	conn.lastUsed = p.timeNow()
	// note the connection is already marked as "in use"
	p.addrConns[conn.addrIndex].conns[conn.connIndex] = conn
	return conn, nil
}
