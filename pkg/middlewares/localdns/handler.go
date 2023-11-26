package localdns

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/local"
	"github.com/qdm12/dns/v2/internal/server"
)

type handler struct {
	// Injected from middleware
	logger Logger
	next   dns.Handler

	// Internal fields
	localExchanges []server.Exchange
	localResolvers []string        // for error messages only
	ctx            context.Context //nolint:containedctx
	cancel         context.CancelFunc
	stopped        atomic.Bool
	waitGroup      sync.WaitGroup
}

func newHandler(resolvers []netip.AddrPort, logger Logger,
	next dns.Handler) *handler {
	dialer := &net.Dialer{
		Timeout: time.Second,
	}
	localExchanges := make([]server.Exchange, len(resolvers))
	localResolvers := make([]string, len(resolvers))
	for i, resolver := range resolvers {
		// WARNING: make sure to pin resolver.String()
		// to a variable for the dial function below!
		resolverAddress := resolver.String()
		localResolvers[i] = resolverAddress
		exchangeName := "local DNS " + resolverAddress
		dial := func(ctx context.Context, _ string, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, "udp", resolverAddress)
		}
		localExchanges[i] = server.NewExchange(
			exchangeName, dial, logger)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &handler{
		ctx:            ctx,
		cancel:         cancel,
		logger:         logger,
		next:           next,
		localExchanges: localExchanges,
		localResolvers: localResolvers,
	}
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

	for i, localExchange := range h.localExchanges {
		response, err := localExchange(h.ctx, r)
		if err != nil {
			h.logger.Debug(err.Error())
			continue
		}

		if response.Rcode != dns.RcodeSuccess {
			h.logger.Debug(fmt.Sprintf(
				"response received for %s from %s has rcode %s",
				r.Question[0].Name, h.localResolvers[i],
				dns.RcodeToString[response.Rcode]))
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

func (h *handler) stop() {
	previouslyStopped := h.stopped.Swap(true)
	if previouslyStopped {
		return
	}

	h.cancel()
	h.waitGroup.Wait()
}
