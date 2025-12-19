package pool

import (
	"context"
	"net"
	"time"
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
type Metrics interface {
	// LiveConnInc increments the number of live connections detected
	// in the pool. This is used when connections are created in [Pool.newConn].
	// You can obtain the current number of live connections by calculating:
	// live_current_conns = live_conn_counter - dead_conn_counter
	LiveConnInc(address string)
	// DeadConnInc increments the number of dead connections detected
	// in the pool. This is used when connections are marked as dead and when
	// dead connections are either removed from the pool or renewed.
	// You can obtain the current number of dead connections by calculating:
	// dead_current_conns = dead_conn_counter - removed_conn_counter
	// - renew_connection_counter{reason="[renewReasonMarkedDead]"}
	DeadConnInc(address string)
	// RemovedConnsAdd adds the given number of removed connections from the pool.
	// This is only called when connections are removed from the pool in [Pool.Put]
	// or [Pool.PutDead].
	RemovedConnsAdd(address string, removed uint)
	// GetConnInc increments the number of connections gotten from the pool.
	// This does not account for failed Get calls, such as dial errors.
	// Tip: to obtain the current in-use connections, use
	// pool_get_connection[success] - pool_put_conns
	GetConnInc(address, outcome string)
	// PutConnInc increments the number of connections put back to the pool.
	// The state argument indicates the state of the connection being put back,
	// "[connStateLive]" (using [Pool.Put]) or "[connStateDead]" (using [Pool.PutDead]).
	PutConnInc(address, state string)
	// NewConnsInc increments the number of new connections created in the pool,
	// even if this is a renewal of a connection slot.
	NewConnsInc(address, outcome string)
	// RenewConnInc increments the counter of renew connection requests.
	RenewConnInc(address, reason, outcome string)
	// RecordUseTime records the duration a connection was in successful use
	// between a [Pool.Get] or [Pool.Renew] and a [Pool.Put].
	// It does not account for connections taken from the pool which failed to
	// be used.
	RecordUseTime(address string, duration time.Duration)
}
