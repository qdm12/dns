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
