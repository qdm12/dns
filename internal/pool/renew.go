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
	address := p.addressFromConn(poolConn)
	now := p.timeNow()
	lifetime := now.Sub(poolConn.created)
	// Ensure this slot cannot be reused while the renewal dials outside the lock.
	poolConn.inUse = true
	p.setConn(poolConn)
	p.mutex.Unlock()

	p.metrics.RecordLifetime(address, lifetime)
	poolConn, err = p.renew(ctx, poolConn, network, renewReasonConnError)
	if err != nil {
		p.metrics.DeadConnInc(address)
		return nil, err
	}
	return poolConn, nil
}

// renew creates a new connection to replace the given one in the pool.
// It dials outside the pool mutex and only locks to mutate pool state.
func (p *Pool) renew(ctx context.Context, conn poolConn, network, reason string) (
	newConn poolConn, err error,
) {
	_ = conn.Close() // ignore error since it may already be closed.

	p.mutex.Lock()
	address := p.addressFromConn(conn)
	p.mutex.Unlock()

	defer func() {
		outcome := outcomeSuccess
		if err != nil {
			outcome = outcomeError
		}
		p.metrics.NewConnsInc(address, outcome)
		p.metrics.RenewConnInc(address, reason, outcome)
	}()

	netConn, err := p.dialer.Dial(ctx, network, address)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if conn.addrIndex >= len(p.addrConns) {
		// The pool was reconfigured/emptied while we were dialing.
		// Drop the new conn and report the original error (or a fresh one).
		if netConn != nil {
			_ = netConn.Close()
		}
		if err == nil {
			err = fmt.Errorf("pool address index %d out of range", conn.addrIndex)
		}
		return poolConn{}, err
	}
	addrConns := p.addrConns[conn.addrIndex]
	p.ensureConnIDToIndex(&addrConns)
	connIndex, found := addrConns.connIDToIndex[conn.id]
	if !found {
		if err != nil {
			return poolConn{}, err
		}
		// The original slot disappeared while dialing. Append a replacement conn
		// and keep it in use for the caller.
		now := p.timeNow()
		connID := p.nextID()
		conn = poolConn{
			Conn:      netConn,
			id:        connID,
			addrIndex: conn.addrIndex,
			connIndex: len(addrConns.conns),
			inUse:     true,
			created:   now,
			lastUsed:  now,
		}
		addrConns.conns = append(addrConns.conns, conn)
		addrConns.connIDToIndex[conn.id] = conn.connIndex
		p.addrConns[conn.addrIndex] = addrConns
		return conn, nil
	}

	conn = addrConns.conns[connIndex]
	conn.connIndex = connIndex
	if err != nil {
		// The pool will retry renewing the connection on another Get call.
		conn.inUse = false
		conn.dead = true
		addrConns.conns[connIndex] = conn
		p.addrConns[conn.addrIndex] = addrConns
		return poolConn{}, err
	}
	conn.Conn = netConn
	conn.dead = false
	now := p.timeNow()
	conn.created = now
	conn.lastUsed = now
	// note the connection is already marked as "in use"
	addrConns.conns[connIndex] = conn
	p.addrConns[conn.addrIndex] = addrConns
	return conn, nil
}
