package dot

import (
	"context"
	"sync"

	"github.com/miekg/dns"
	"github.com/qdm12/golibs/logging"
)

type handler struct {
	// External objects
	ctx    context.Context
	logger logging.Logger

	// Internal objects
	dial          dialFunc
	udpBufferPool *sync.Pool
	client        *dns.Client
}

func newDNSHandler(ctx context.Context, logger logging.Logger,
	options ...Option) dns.Handler {
	settings := defaultSettings()
	for _, option := range options {
		option(&settings)
	}

	const dnsPacketMaxSize = 512
	udpBufferPool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, dnsPacketMaxSize)
		},
	}

	return &handler{
		ctx:           ctx,
		logger:        logger,
		dial:          newDoTDial(settings),
		udpBufferPool: udpBufferPool,
		client:        &dns.Client{},
	}
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	DoTConn, err := h.dial(h.ctx, "", "")
	if err != nil {
		h.logger.Warn("cannot dial: %s", err)
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}
	conn := &dns.Conn{Conn: DoTConn}

	response, _, err := h.client.ExchangeWithConn(r, conn)

	if err := conn.Close(); err != nil {
		h.logger.Warn("cannot close the DoT connection: %s", err)
	}

	if err != nil {
		h.logger.Warn("cannot exchange over DoT connection: %s", err)
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}

	if err := w.WriteMsg(response); err != nil {
		h.logger.Warn("cannot write DNS message back to client: %s", err)
	}
}
