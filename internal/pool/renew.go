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
	address := p.addressFromConn(poolConn)
	p.metrics.RenewRequestsInc(address)
	p.metrics.LiveConnsAdd(address, -1)
	// note: no need to lock earlier
	p.mutex.Lock()
	defer p.mutex.Unlock()
	poolConn, err = p.renew(ctx, poolConn, network)
	if err != nil {
		return nil, err
	}
	p.addrConns[poolConn.addrIndex].conns[poolConn.connIndex] = poolConn
	return poolConn, nil
}

func (p *Pool) renew(ctx context.Context, conn poolConn, network string) (newConn poolConn, err error) {
	_ = conn.Close() // ignore error since it may already be closed.
	address := p.addressFromConn(conn)
	p.metrics.RenewalsInc(address)
	netConn, err := p.dialer.Dial(ctx, network, address)
	if err != nil {
		// The pool will retry renewing the connection on another Get call.
		return poolConn{}, err
	}
	conn.Conn = netConn
	conn.dead = false
	// note the connection is already marked as "in use"
	p.metrics.LiveConnsAdd(address, 1)
	return conn, nil
}
