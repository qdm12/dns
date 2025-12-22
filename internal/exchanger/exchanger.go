package exchanger

import (
	"context"
	"errors"
	"fmt"
	"hash/maphash"
	"io"
	"math/rand/v2"
	"net"
	"strings"
	"syscall"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/pool"
)

type Exchanger struct {
	client     *dns.Client
	dialer     Dialer
	warner     Warner
	reuseConns bool
	pool       *pool.Pool
	rand       *rand.Rand
	addresses  []string
}

func New(dialer Dialer, poolMetrics PoolMetrics, warner Warner) *Exchanger {
	reuseConns := dialer.ReusableConnsSupported()
	addresses := dialer.Addresses()
	if len(addresses) == 0 {
		panic("dialer " + dialer.String() + " has no addresses")
	}
	return &Exchanger{
		client:     &dns.Client{},
		dialer:     dialer,
		warner:     warner,
		reuseConns: reuseConns,
		pool:       pool.New(dialer, poolMetrics),
		rand:       rand.New(newMaphashSource()), //nolint:gosec
		addresses:  addresses,
	}
}

var ErrDialFailed = errors.New("dial failed")

func (e *Exchanger) Exchange(ctx context.Context, network string, request *dns.Msg) (
	response *dns.Msg, err error,
) {
	if e.reuseConns {
		return e.exchangeWithPool(ctx, network, request) // dot, doh
	}
	return e.exchangeWithRand(ctx, network, request) // plain
}

func (e *Exchanger) exchangeWithPool(ctx context.Context, network string, request *dns.Msg) (
	response *dns.Msg, err error,
) {
	netConn, err := e.pool.Get(ctx, network)
	if err != nil {
		return nil, fmt.Errorf("getting %s connection for request %s: %w",
			e.dialer, extractRequestQuestion(request), err)
	}

	dnsConn := &dns.Conn{Conn: netConn}
	response, roundTripDuration, err := e.client.ExchangeWithConnContext(ctx, request, dnsConn)
	if err == nil {
		e.pool.Put(netConn)
		return response, nil
	}

	if !isClosedConnErr(err) {
		e.pool.PutDead(netConn)
		roundTripMilliseconds := roundTripDuration.Round(time.Millisecond).Milliseconds()
		return nil, fmt.Errorf("exchanging over %s connection (%dms) for request %s: %w",
			e.dialer, roundTripMilliseconds, extractRequestQuestion(request), err)
	}

	// Connection is closed, try to renew it
	_ = dnsConn.Close()
	netConn, err = e.pool.Renew(ctx, network, netConn)
	if err != nil {
		return nil, fmt.Errorf("renewing %s connection for request %s: %w",
			e.dialer, extractRequestQuestion(request), err)
	}
	dnsConn = &dns.Conn{Conn: netConn}
	response, roundTripDuration, err = e.client.ExchangeWithConnContext(ctx, request, dnsConn)
	if err != nil {
		e.pool.PutDead(netConn)
		roundTripMilliseconds := roundTripDuration.Round(time.Millisecond).Milliseconds()
		return nil, fmt.Errorf("exchanging over %s connection (%dms) for request %s: %w",
			e.dialer, roundTripMilliseconds, extractRequestQuestion(request), err)
	}

	e.pool.Put(netConn)
	return response, nil
}

func (e *Exchanger) exchangeWithRand(ctx context.Context, network string, request *dns.Msg) (
	response *dns.Msg, err error,
) {
	addrOrURL := e.addresses[e.rand.IntN(len(e.addresses))]
	netConn, err := e.dialer.Dial(ctx, network, addrOrURL)
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

func isClosedConnErr(err error) bool {
	return errors.Is(err, net.ErrClosed) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ECONNRESET)
}

func newMaphashSource() *mapHashSource {
	return &mapHashSource{}
}

type mapHashSource struct{}

func (s *mapHashSource) Uint64() uint64 {
	return new(maphash.Hash).Sum64()
}
