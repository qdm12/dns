package dot

import (
	"context"
	"fmt"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/log"
)

type exchangeFunc func(ctx context.Context, request *dns.Msg) (response *dns.Msg, err error)

func makeDNSExchange(client *dns.Client, dial dialFunc, warner log.Warner) exchangeFunc {
	return func(ctx context.Context, request *dns.Msg) (response *dns.Msg, err error) {
		netConn, err := dial(ctx, "", "")
		if err != nil {
			return nil, fmt.Errorf("cannot dial DoT server: %w", err)
		}
		dnsConn := &dns.Conn{Conn: netConn}

		response, _, err = client.ExchangeWithConn(request, dnsConn)

		if closeErr := dnsConn.Close(); closeErr != nil {
			warner.Warn("cannot close DoT connection: " + closeErr.Error())
		}

		if err != nil {
			return nil, fmt.Errorf("cannot exchange over DoT connection: %w", err)
		}

		return response, nil
	}
}
