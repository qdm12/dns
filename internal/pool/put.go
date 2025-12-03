package pool

import (
	"fmt"
	"net"
	"time"
)

// Put puts back a working connection to the pool.
// If the connection is no longer working, use either [pool.Renew] to replace
// the connection or use [pool.PutDead] if you don't want to renew it.
// This function cleans up dead connections from the pool, if possible.
func (p *Pool) Put(conn net.Conn) {
	poolConn, ok := conn.(poolConn)
	if !ok {
		panic(fmt.Sprintf("cannot put back non-pool connection %T", conn))
	}

	poolConn.lastUsed = p.timeNow()
	poolConn.inUse = false

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.addrConns[poolConn.addrIndex].conns[poolConn.connIndex] = poolConn

	p.cleanup(poolConn.addrIndex)

	address := p.addressFromConn(poolConn)
	p.metrics.InUseConnsAdd(address, -1)
}

func (p *Pool) PutDead(conn net.Conn) {
	poolConn, ok := conn.(poolConn)
	if !ok {
		panic(fmt.Sprintf("cannot put back dead non-pool connection %T", conn))
	}

	poolConn.inUse = false
	poolConn.dead = true
	poolConn.lastUsed = time.Now()

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.addrConns[poolConn.addrIndex].conns[poolConn.connIndex] = poolConn

	p.cleanup(poolConn.addrIndex)

	address := p.addressFromConn(poolConn)
	p.metrics.InUseConnsAdd(address, -1)
	p.metrics.LiveConnsAdd(address, -1)
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

	address := p.addrConns[addrIndex].address
	p.metrics.ConnsAdd(address, -removed)
	p.addrConns[addrIndex].conns = conns
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
func (p *Pool) compact(conns []poolConn) (updated []poolConn, removed int) {
	// writeIndex tracks the next available slot for a movable/alive connection.
	writeIndex := 0

	for readIndex, conn := range conns {
		switch {
		case p.isConnDead(conn):
			// Skip this connection.
			// writeIndex stays pointing to it, ready to be overwritten/swapped
			// with the next alive unused connection.
			continue
		case conn.inUse:
			// The connection is in use and cannot be moved.
			// Ensure the compaction frontier writeIndex skips it ONLY if it
			// is blocking the frontier.
			if writeIndex == readIndex {
				writeIndex++
			}
			continue
		}

		// Handle the movable alive and unused connection: compact it to writeIndex.

		// Move the writeIndex forward until it points to a movable connection
		for writeIndex < readIndex && conns[writeIndex].inUse {
			writeIndex++
		}

		if writeIndex == readIndex {
			// The connection is at the correct position at the frontier, so advance the frontier.
			writeIndex++
			continue
		}
		// Swap the alive unused connection pointed by readIdx with the dead connection
		// pointed by writeIdx.
		conns[readIndex], conns[writeIndex] = conns[writeIndex], conns[readIndex]
		conns[readIndex].connIndex = readIndex
		conns[writeIndex].connIndex = writeIndex
		writeIndex++ // advance the compaction frontier
	}

	finalLength := 0
	for i := len(conns) - 1; i >= 0; i-- {
		if !conns[i].dead {
			finalLength = i + 1
			break
		}
	}

	return conns[:finalLength], len(conns) - finalLength
}
