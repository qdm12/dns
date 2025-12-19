package pool

import (
	"context"
	"fmt"
	"math"
	"net"
)

func (p *Pool) Get(ctx context.Context, network string) (netConn net.Conn, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var address string
	defer func() {
		outcome := outcomeSuccess
		if err != nil {
			outcome = outcomeError
		}
		p.metrics.GetConnInc(address, outcome)
	}()

	if !p.oneConnPerAddr {
		// Not all addresses have one connection yet, so we need
		// to create a new connection for any address without one.
		// This is a form of lazy loading for creating one connection
		// per address.
		conn, addrIndex, err := p.newConn(ctx, network)
		address = p.addrConns[addrIndex].address
		if err != nil {
			return nil, fmt.Errorf("creating connection: %w", err)
		}
		p.setIfAllAddrsHaveOneConn()
		return conn, nil
	}

	conn, found, live := p.findNextAvailConn()
	if !found {
		// All addresses have one connection, but all connections
		// are in use, so we need to create an additional connection.
		conn, addrIndex, err := p.newConn(ctx, network)
		address = p.addrConns[addrIndex].address
		if err != nil {
			return nil, fmt.Errorf("creating connection: %w", err)
		}
		return conn, nil
	}

	// Even if the connection fails to renew, let's increase lastUsedAddrIndex
	// to avoid always trying the same address, in case it becomes inaccessible.
	p.lastUsedAddrIndex = conn.addrIndex
	conn.inUse = true
	address = p.addressFromConn(conn)
	if live {
		conn.lastUsed = p.timeNow()
		p.addrConns[conn.addrIndex].conns[conn.connIndex] = conn
		return conn, nil
	}
	conn, err = p.renew(ctx, conn, network, "marked dead")
	if err != nil {
		return nil, fmt.Errorf("renewing dead connection: %w", err)
	}
	p.metrics.LiveConnInc(address)
	return conn, nil
}

// findNextAvailConn finds the next available connection in the pool.
// It first checks the connections of the address right after the last used address.
// It favors returning a live connection, but will return a dead
// connection if no live connection is available so it can be renewed.
// If all connections for the address are live and in use, it checks the next
// addresses. If all addresses have all connections live and in use, it returns
// no connection with found set to false, indicating a new connection should be created
// for the initially checked address.
// This is as such to prevent creating more connections for servers with high latency which
// would keep connections in use for longer periods of time.
func (p *Pool) findNextAvailConn() (conn poolConn, found, live bool) {
	startIndex := (p.lastUsedAddrIndex + 1) % len(p.addrConns)
	addrIndex := startIndex
	firstCheck := true
	for {
		if firstCheck {
			firstCheck = false
		} else if addrIndex == startIndex {
			// we are back to the first address checked, no available connections
			return poolConn{}, false, false
		}
		addrConn := p.addrConns[addrIndex]
		deadIndex := -1
		for i, conn := range addrConn.conns {
			if conn.inUse {
				continue
			} else if p.isConnDead(conn) {
				if deadIndex == -1 {
					deadIndex = i
				}
				continue
			}
			// Prefer using a live available connection instead of renewing a dead one
			return conn, true, true
		}
		if deadIndex != -1 {
			// Return the first dead connection found to be renewed
			return addrConn.conns[deadIndex], true, false
		}
		// All connections are in use for this address, try the next one
		addrIndex = (addrIndex + 1) % len(p.addrConns)
	}
}

// newConn adds a new connection to the pool using either the address
// with the least connections already in the pool, or the "next" address
// without any connections, and returns it.
// It also returns the address index, even on error, for metrics usage in
// the calling code.
func (p *Pool) newConn(ctx context.Context, network string) (
	conn poolConn, addrIndex int, err error,
) {
	index := p.findAddressForNewConn()
	addrConns := p.addrConns[index]
	address := addrConns.address

	defer func() {
		if err != nil {
			p.metrics.NewConnsInc(address, outcomeError)
		} else {
			p.metrics.NewConnsInc(address, outcomeSuccess)
			p.metrics.LiveConnInc(address)
		}
	}()

	netConn, err := p.dialer.Dial(ctx, network, address)
	if err != nil {
		return poolConn{}, index, err
	}
	now := p.timeNow()
	conn = poolConn{
		Conn:      netConn,
		addrIndex: index,
		connIndex: len(addrConns.conns),
		inUse:     true, // about to be returned and used
		created:   now,
		lastUsed:  now,
	}
	addrConns.conns = append(addrConns.conns, conn)
	p.addrConns[index] = addrConns
	p.lastUsedAddrIndex = conn.addrIndex
	return conn, index, nil
}

// findAddressForNewConn finds the address index to use for creating a new connection.
// It prefers addresses without any connections yet, otherwise it returns
// the address with the least number of connections.
// This favors equal parallelism across all addresses.
// It scans round-robin from the next address to the last used address, in
// the order of p.addrConns.
func (p *Pool) findAddressForNewConn() (addressIndex int) {
	minNumberOfConns := math.MaxInt
	minIndex := -1
	index := (p.lastUsedAddrIndex + 1) % len(p.addrConns)
	for range len(p.addrConns) {
		numberOfConns := len(p.addrConns[index].conns)
		switch {
		case numberOfConns == 0:
			return index
		case numberOfConns < minNumberOfConns:
			minNumberOfConns = numberOfConns
			minIndex = index
		}
		index = (index + 1) % len(p.addrConns)
	}
	return minIndex
}
