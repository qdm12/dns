package doh

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/dot"
	"github.com/qdm12/dns/pkg/provider"
	"github.com/qdm12/golibs/logging"
)

type handler struct {
	// Configuration
	dohServers []provider.DoHServer

	// External objects
	ctx    context.Context
	logger logging.Logger

	// Internal objects
	udpBufferPool *sync.Pool
	httpHandler   HTTPHandler
	counter       *int64
}

func newDNSHandler(ctx context.Context, logger logging.Logger,
	options ...Option) dns.Handler {
	settings := defaultSettings()
	for _, option := range options {
		option(&settings)
	}

	dotOptions := []dot.Option{
		dot.Providers(settings.providers[0], settings.providers[1:]...),
		dot.Timeout(settings.timeout),
	}
	if settings.ipv6 {
		dotOptions = append(dotOptions, dot.IPv6())
	}

	const dnsPacketMaxSize = 512
	udpBufferPool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, dnsPacketMaxSize)
		},
	}

	return &handler{
		dohServers:    settings.dohServers,
		ctx:           ctx,
		logger:        logger,
		udpBufferPool: udpBufferPool,
		httpHandler:   newHTTPHandler(dotOptions...),
		counter:       new(int64),
	}
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	buffer := h.udpBufferPool.Get().([]byte)
	// no need to reset buffer as PackBuffer takes care of
	// slicing it down
	wire, err := r.PackBuffer(buffer)
	if err != nil {
		h.logger.Warn("cannot pack DNS message to wire format: %s", err)
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}

	// Pick DoH url pseudo-randomly from the chosen providers
	i := int(atomic.AddInt64(h.counter, 1))
	i %= len(h.dohServers)
	dohServer := h.dohServers[i]

	respWire, err := h.httpHandler.Request(h.ctx, dohServer.URL, wire)

	// It's fine to copy the slice headers as long as we keep
	// the underlying array of bytes.
	h.udpBufferPool.Put(buffer) //nolint:staticcheck

	if err != nil {
		h.logger.Warn("HTTP request failed: %s", err)
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}

	message := new(dns.Msg)
	if err := message.Unpack(respWire); err != nil {
		h.logger.Warn("cannot unpack message from wire format: %s", err)
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}

	message.SetReply(r)
	if err := w.WriteMsg(message); err != nil {
		h.logger.Warn("cannot write DNS message back to client: %s", err)
	}
}
