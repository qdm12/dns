package exchanger

import (
	"context"
	"net"
)

type Dialer interface {
	String() string
	// ReusableConnsSupported returns true if the dialer supports reusing connections.
	ReusableConnsSupported() bool
	// Addresses returns the list of all addresses/URLs of the upstream servers
	// known by this dialer.
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

type Warner interface {
	Warn(s string)
}

type PoolMetrics interface {
	LiveConnInc(address string)
	DeadConnInc(address string)
	RemovedConnsAdd(address string, removed uint)
	GetConnInc(address, outcome string)
	PutConnInc(address, state string)
	NewConnsInc(address, outcome string)
	RenewConnInc(address, reason, outcome string)
}
