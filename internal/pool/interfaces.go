package pool

import (
	"context"
	"net"
)

type Dialer interface {
	// Addresses returns the list of addresses or URLs the dialer can dial to.
	Addresses() []string
	// Dial dials a network connection to the given network and address/url.
	// Note the network is typically "tcp" or "udp", and can be used or ignored.
	// For example, the DoT dialer always dials "tcp" connections, so it ignores
	// the network parameter, and same goes for the DoH dialer.
	// Note addrOrURL is used differently depending on the dialer. For DoT and plain
	// dialers, it is a address:port. For DoH dialers, it is a URL.
	// Finally, if the addrOrURL is not recognized by the dialer implementation, the
	// dialer, notably the DoT, DoH and plain dialers defined in this project,
	// pick a server at random from their list of servers. This is as such so these
	// dialers' Dial method can be used in [net.Resolver] Dial field.
	Dial(ctx context.Context, network, addrOrURL string) (conn net.Conn, err error)
}

// Metrics defines the metric calls made by the connection pool.
// Some metrics can be extrapolated using the following:
// - current live connections = pool_new_connections[success] - pool_dead_connections
// - current dead connections = pool_dead_connections - pool_removed_connections
// - current in-use connections = pool_get_connection[success] - pool_put_connection
// The following methods are all explained in details below.
type Metrics interface {
	// NewConnsInc increments the number of new connections created in the pool,
	// even if this is a renewal of a connection slot.
	// Tip: to obtain the current live connections, use
	// pool_new_connections[success] - pool_dead_connections
	NewConnsInc(address, outcome string)
	// RenewedConnsInc increments the counter of renewed connections.
	// Note the renewed connections count is a subset of the new connections count.
	RenewedConnsInc(address, reason, outcome string)
	// DeadConnsInc increments the number of dead connections detected in the pool,
	// either by the caller which would use [pool.PutDead] or by the pool itself either
	// when marking connections as dead in [pool.Get] or right before removing
	// connections when [pool.Put] or [pool.PutDead] is called.
	// Tip: to obtain the current number of dead connections, use
	// pool_dead_connections - pool_removed_connections
	DeadConnsInc(address string)
	// RemovedConnsAdd adds the given number of removed connections from the pool.
	// This is only called when connections are removed from the pool in [pool.Put]
	// or [pool.PutDead].
	RemovedConnsAdd(address string, removed uint)
	// GetConnsInc increments the number of connections gotten from the pool.
	// This does not account for failed Get calls, such as dial errors.
	// Tip: to obtain the current in-use connections, use
	// pool_get_connection[success] - pool_put_conns
	GetConnsInc(address, outcome string)
	// PutConnsInc increments the number of connections put back to the pool.
	// The state argument indicates the state of the connection being put back,
	// "live" (using [pool.Put]) or "dead" (using [pool.PutDead]).
	PutConnsInc(address, state string)
}
