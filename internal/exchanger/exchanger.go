package exchanger

import (
	"context"
	"fmt"
	"strings"
	"time"

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
		return nil, fmt.Errorf("dialing %s server for request %s: %w",
			e.dialer, extractRequestQuestion(request), err)
	}
	dnsConn := &dns.Conn{Conn: netConn}

	response, roundTripDuration, err := e.client.ExchangeWithConnContext(ctx, request, dnsConn)

	closeErr := dnsConn.Close()
	if closeErr != nil {
		e.warner.Warn("cannot close " + e.dialer.String() + " connection: " + closeErr.Error())
	}

	if err != nil {
		roundTripMilliseconds := roundTripDuration.Round(time.Millisecond).Milliseconds()
		return nil, fmt.Errorf("exchanging over %s connection (%dms) for request %s: %w",
			e.dialer, roundTripMilliseconds, extractRequestQuestion(request), err)
	}

	return response, nil
}

func extractRequestQuestion(request *dns.Msg) (s string) {
	if len(request.Question) == 0 {
		return "<no question>"
	}
	question := request.Question[0]
	return dns.ClassToString[question.Qclass] + " " +
		dns.TypeToString[question.Qtype] + " " +
		strings.ToLower(question.Name)
}
