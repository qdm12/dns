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

type Metrics interface {
	ConnsAdd(address string, n int)
	LiveConnsAdd(address string, n int)
	InUseConnsAdd(address string, n int)
	RenewRequestsInc(address string)
	RenewalsInc(address string)
}
