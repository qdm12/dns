package doh

import (
	"context"

	"github.com/miekg/dns"
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
	dial   dialFunc
	client *dns.Client
}

func newDNSHandler(ctx context.Context, logger logging.Logger,
	options ...Option) dns.Handler {
	settings := defaultSettings()
	for _, option := range options {
		option(&settings)
	}

	return &handler{
		dohServers: settings.dohServers,
		ctx:        ctx,
		logger:     logger,
		dial:       newDoHDial(settings),
		client:     &dns.Client{},
	}
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	DoHConn, err := h.dial(h.ctx, "", "")
	if err != nil {
		h.logger.Warn("cannot dial: %s", err)
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}

	conn := &dns.Conn{Conn: DoHConn}

	response, _, err := h.client.ExchangeWithConn(r, conn)
	if err != nil {
		h.logger.Warn("cannot exchange over DoH connection: %s", err)
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}

	response.SetReply(r)
	if err := w.WriteMsg(response); err != nil {
		h.logger.Warn("cannot write DNS message back to client: %s", err)
	}
}
