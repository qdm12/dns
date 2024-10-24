package exchanger

import (
	"context"
	"fmt"

	"github.com/miekg/dns"
)

type Exchanger struct {
	client *dns.Client
	dialer Dialer
	warner Warner
}

func New(dialer Dialer, warner Warner) *Exchanger {
	return &Exchanger{
		client: &dns.Client{},
		dialer: dialer,
		warner: warner,
	}
}

func (e *Exchanger) Exchange(ctx context.Context, network string, request *dns.Msg) (
	response *dns.Msg, err error,
) {
	netConn, err := e.dialer.Dial(ctx, network, "")
	if err != nil {
		return nil, fmt.Errorf("dialing %s server: %w", e.dialer, err)
	}
	dnsConn := &dns.Conn{Conn: netConn}

	response, _, err = e.client.ExchangeWithConnContext(ctx, request, dnsConn)

	closeErr := dnsConn.Close()
	if closeErr != nil {
		e.warner.Warn("cannot close " + e.dialer.String() + " connection: " + closeErr.Error())
	}

	if err != nil {
		return nil, fmt.Errorf("exchanging over %s connection: %w", e.dialer, err)
	}

	return response, nil
}
