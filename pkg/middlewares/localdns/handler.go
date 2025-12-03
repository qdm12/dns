package localdns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/exchanger"
	"github.com/qdm12/dns/v2/internal/local"
)

type handler struct {
	// Injected from middleware
	logger      Logger
	next        dns.Handler
	timeoutWarn bool

	// Internal fields
	localExchanges []exchangerIntf
	localResolvers []string        // for error messages only
	ctx            context.Context //nolint:containedctx
	cancel         context.CancelFunc
	stopped        atomic.Bool
	waitGroup      sync.WaitGroup
}

func newHandler(resolvers []netip.AddrPort, logger Logger,
	next dns.Handler, timeoutWarn bool,
) *handler {
	netDialer := &net.Dialer{
		Timeout: time.Second,
	}
	localExchangers := make([]exchangerIntf, len(resolvers))
	localResolvers := make([]string, len(resolvers))
	for i, resolver := range resolvers {
		// WARNING: make sure to pin resolver.String()
		// to a variable for the dial function below!
		resolverAddress := resolver.String()
		localResolvers[i] = resolverAddress
		dialer := &dialer{
			netDialer:       netDialer,
			resolverAddress: resolverAddress,
		}
		poolMetrics := (exchanger.PoolMetrics)(nil) // pool is not used
		localExchangers[i] = exchanger.New(dialer, poolMetrics, logger)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &handler{
		ctx:            ctx,
		cancel:         cancel,
		logger:         logger,
		next:           next,
		timeoutWarn:    timeoutWarn,
		localExchanges: localExchangers,
		localResolvers: localResolvers,
	}
}

type dialer struct {
	netDialer       *net.Dialer
	resolverAddress string
}

func (d *dialer) Dial(ctx context.Context, network, _ string) (net.Conn, error) {
	return d.netDialer.DialContext(ctx, network, d.resolverAddress)
}

func (d *dialer) String() string {
	return "local DNS " + d.resolverAddress
}

func (d *dialer) ReusableConnsSupported() bool {
	return false
}

func (d *dialer) Addresses() []string {
	return []string{d.resolverAddress}
}

// ServeDNS implements the dns.Handler interface for the
// localdns middleware handler.
// It redirects DNS requests containing a single local
// name question to the local DNS servers specified.
func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if h.stopped.Load() {
		return
	}
	h.waitGroup.Add(1)
	defer h.waitGroup.Done()

	// This middleware only handles single question requests
	// with a local name question. If there is no question or
	// more than one question, we just pass the request through
	// to the next handler.
	// Usually DNS requests only have a single question,
	// see https://github.com/miekg/dns/issues/396#issuecomment-240149439
	const expectedQuestionCount = 1
	if len(r.Question) != expectedQuestionCount ||
		!local.IsFQDNLocal(r.Question[0].Name) {
		h.next.ServeDNS(w, r)
		return
	}

	r.Question[0].Name = strings.ToLower(r.Question[0].Name)

	for i, localExchange := range h.localExchanges {
		response, err := h.tryExchange(localExchange, h.localResolvers[i], r)
		if err != nil {
			if errors.Is(err, errRcodeNotSuccess) {
				continue // try next resolver
			}
			if h.timeoutWarn || !strings.HasSuffix(err.Error(), ": i/o timeout") {
				h.logger.Warn(err.Error())
			} else {
				h.logger.Debug(err.Error())
			}
			continue
		}
		_ = w.WriteMsg(response)
		return
	}

	response := new(dns.Msg)
	response.SetReply(r)
	response.SetRcode(r, dns.RcodeNameError)
	_ = w.WriteMsg(response)
}

var errRcodeNotSuccess = errors.New("rcode is not success")

func (h *handler) tryExchange(exchange exchangerIntf, resolverName string, r *dns.Msg) (
	response *dns.Msg, err error,
) {
	response, err = exchange.Exchange(h.ctx, "udp", r)
	if err != nil {
		return nil, fmt.Errorf("exchanging over udp: %w", err)
	}

	if response.Rcode != dns.RcodeSuccess {
		if response.Rcode != dns.RcodeNameError {
			// Do not debug log success responses or name error responses since these happen
			// often, notably due to /etc/resolv.conf search domains.
			h.logger.Debug(fmt.Sprintf(
				"response received for %s from %s over udp has rcode %s",
				r.Question[0].Name, resolverName,
				dns.RcodeToString[response.Rcode]))
		}
		return nil, fmt.Errorf("%w", errRcodeNotSuccess)
	}

	if response.Truncated {
		response, err := exchange.Exchange(h.ctx, "tcp", r)
		if err != nil {
			return nil, fmt.Errorf("exchanging over tcp: %w", err)
		}

		if response.Rcode != dns.RcodeSuccess {
			h.logger.Debug(fmt.Sprintf(
				"response received for %s from %s over tcp has rcode %s",
				r.Question[0].Name, resolverName,
				dns.RcodeToString[response.Rcode]))
			return nil, fmt.Errorf("%w", errRcodeNotSuccess)
		}
	}

	return response, nil
}

func (h *handler) stop() {
	previouslyStopped := h.stopped.Swap(true)
	if previouslyStopped {
		return
	}

	h.cancel()
	h.waitGroup.Wait()
}
