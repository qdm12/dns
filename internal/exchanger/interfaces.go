package exchanger

import (
	"context"
	"net"
)

type Dialer interface {
	String() string
	Dial(ctx context.Context, network, _ string) (conn net.Conn, err error)
}

type Warner interface {
	Warn(s string)
}
