package server

import (
	"context"
	"fmt"
	"net"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/log"
)

type Exchange = func(ctx context.Context, request *dns.Msg) (
	response *dns.Msg, err error,
)

type Dial = func(ctx context.Context, _, _ string) (net.Conn, error)

func NewExchange(name string, dial Dial, warner log.Warner) Exchange {
	client := &dns.Client{}
	return func(ctx context.Context, request *dns.Msg) (response *dns.Msg, err error) {
		netConn, err := dial(ctx, "", "")
		if err != nil {
			return nil, fmt.Errorf("dialing %s server: %w", name, err)
		}
		dnsConn := &dns.Conn{Conn: netConn}

		response, _, err = client.ExchangeWithConn(request, dnsConn)

		if closeErr := dnsConn.Close(); closeErr != nil {
			warner.Warn("cannot close " + name + " connection: " + closeErr.Error())
		}

		if err != nil {
			return nil, fmt.Errorf("exchanging over %s connection: %w", name, err)
		}

		return response, nil
	}
}
