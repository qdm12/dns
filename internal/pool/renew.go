package pool

import (
	"context"
	"fmt"
	"net"
)

// Renew creates a new connection to replace the given one in the pool.
// It should be used when the caller detects that the given connection is dead,
// for example after receiving a io.EOF error.
func (p *Pool) Renew(ctx context.Context, network string, conn net.Conn) (newConn net.Conn, err error) {
	poolConn, ok := conn.(poolConn)
	if !ok {
		panic(fmt.Sprintf("cannot renew non-pool connection %T", conn))
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	const reason = "connection error"
	poolConn, err = p.renew(ctx, poolConn, network, reason)
	if err != nil {
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
	netConn, err := p.dialer.Dial(ctx, network, address)
	if err != nil {
		// The pool will retry renewing the connection on another Get call.
		conn.inUse = false
		conn.dead = true
		p.addrConns[conn.addrIndex].conns[conn.connIndex] = conn
		p.metrics.DeadConnsInc(address)
		p.metrics.NewConnsInc(address, outcomeError)
		p.metrics.RenewedConnsInc(address, reason, outcomeError)
		// Note: return the original conn marked as dead since it contains
		// the address index which is used for metrics.
		return conn, err
	}
	conn.Conn = netConn
	conn.dead = false
	// note the connection is already marked as "in use"
	p.addrConns[conn.addrIndex].conns[conn.connIndex] = conn
	p.metrics.NewConnsInc(address, outcomeSuccess)
	p.metrics.RenewedConnsInc(address, reason, outcomeSuccess)
	return conn, nil
}
