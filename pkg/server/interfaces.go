package server

import (
	"context"
	"net"

	"github.com/miekg/dns"
)

type Dialer interface {
	Dial(ctx context.Context, network, _ string) (conn net.Conn, err error)
	// String returns the connection type the dialer will use, for example
	// "dns over tls".
	String() string
	// ReusableConnsSupported returns true if the dialer supports reusing connections.
	ReusableConnsSupported() bool
	// Addresses returns the list of all addresses/URLs of the upstream servers
	// known by this dialer.
	Addresses() []string
}

type PoolMetrics interface {
	ConnsAdd(address string, n int)
	LiveConnsAdd(address string, n int)
	InUseConnsAdd(address string, n int)
	RenewRequestsInc(address string)
	RenewalsInc(address string)
}

type Middleware interface {
	String() string
	Wrap(next dns.Handler) dns.Handler
	Stop() (err error)
}

type Logger interface {
	Debug(s string)
	Info(s string)
	Warn(s string)
}
