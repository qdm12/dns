package pool

import (
	"fmt"
	"net"
)

const (
	connStateLive = "live"
	connStateDead = "dead"
)

// Put puts back a working connection to the pool.
// If the connection is no longer working, use either [Pool.Renew] to replace
// the connection or use [Pool.PutDead] if you don't want to renew it,
// to indicate to the pool it is dead.
func (p *Pool) Put(conn net.Conn) {
	poolConn, ok := conn.(poolConn)
	if !ok {
		panic(fmt.Sprintf("cannot put back non-pool connection %T", conn))
	}

	now := p.timeNow()
	inUseDuration := now.Sub(poolConn.lastUsed)
	address := p.addressFromConn(poolConn)
	p.metrics.RecordUseTime(address, inUseDuration)

	poolConn.lastUsed = now
	poolConn.inUse = false

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.addrConns[poolConn.addrIndex].conns[poolConn.connIndex] = poolConn

	p.cleanup(poolConn.addrIndex)

	p.metrics.PutConnInc(address, connStateLive)
}

// PutDead must be called instead of [Pool.Put] to put back a dead connection
// to the pool, EXCEPT in the following two cases:
//   - after a failed call to [Pool.Get], in which case the connection is nil and
//     this would panic.
//   - after a failed call to [Pool.Renew], in which case the connection is already
//     marked as dead
func (p *Pool) PutDead(conn net.Conn) {
	poolConn, ok := conn.(poolConn)
	if !ok {
		panic(fmt.Sprintf("cannot put back dead non-pool connection %T", conn))
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	isMarkedDead := p.addrConns[poolConn.addrIndex].conns[poolConn.connIndex].dead
	if isMarkedDead {
		// Just in case the caller calls this function after a failed [Pool.Renew]
		return
	}

	poolConn.inUse = false
	poolConn.dead = true // do not update metrics in [Pool.cleanup]
	address := p.addressFromConn(poolConn)
	now := p.timeNow()
	lifetime := now.Sub(poolConn.created)
	p.metrics.RecordLifetime(address, lifetime)
	poolConn.lastUsed = now

	p.addrConns[poolConn.addrIndex].conns[poolConn.connIndex] = poolConn
	p.cleanup(poolConn.addrIndex)

	p.metrics.PutConnInc(address, connStateDead)
	p.metrics.DeadConnInc(address)
}

// cleanup cleans up connections for the given address index.
func (p *Pool) cleanup(addrIndex int) {
	conns := p.addrConns[addrIndex].conns
	if len(conns) == 1 {
		// In case we return a single dead connection, keep it, the next call to
		// [Pool.Get] would renew it.
		return
	}

	conns, removed := p.compact(conns)
	p.addrConns[addrIndex].conns = conns
	if removed > 0 {
		address := p.addrConns[addrIndex].address
		p.metrics.RemovedConnsAdd(address, removed)
	}
}

// compact aims to reduce the connections slice length by removing
// dead connections. To do so, it:
//   - does not move in use connections, because their index is in use by the caller
//   - moves dead connections to the "right" side by swapping alive unused connections
//     with dead connections
//   - removes the sequence of dead connections from the end of the slice
//
// Since this is run when putting back a connection in the pool, it must be efficient.
// The complexity is O(n) because each item is visited by readIndex and
// potentially visited by writeIndex at most once.
// The space complexity is O(1) as it operates entirely in-place using swaps.
func (p *Pool) compact(conns []poolConn) (updated []poolConn, removed uint) {
	// writeIndex tracks the next available slot for a movable/alive connection.
	writeIndex := 0

	for readIndex, conn := range conns {
		switch {
		case conn.inUse: // cannot be dead
			// The connection is in use and cannot be moved.
			// Ensure the compaction frontier writeIndex skips it ONLY if it
			// is blocking the frontier.
			if writeIndex == readIndex {
				writeIndex++
			}
			continue
		case p.isConnDead(conn):
			// writeIndex stays pointing to it, ready to be overwritten/swapped
			// with the next alive unused connection.
			continue
		case writeIndex == readIndex:
			// The connection is at the correct position at the frontier, so advance the frontier.
			writeIndex++
			continue
		}

		// Swap the alive unused connection pointed by readIdx with the dead connection
		// pointed by writeIdx.
		conns[readIndex], conns[writeIndex] = conns[writeIndex], conns[readIndex]
		conns[readIndex].connIndex = readIndex
		conns[writeIndex].connIndex = writeIndex
		writeIndex++
	}

	finalLength := 0
	for i := len(conns) - 1; i >= 0; i-- {
		if !conns[i].dead {
			finalLength = i + 1
			break
		}
	}

	removed = uint(len(conns) - finalLength) //nolint:gosec
	return conns[:finalLength], removed
}
