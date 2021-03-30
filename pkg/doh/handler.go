package doh

import (
	"context"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/dns/pkg/cache"
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
	cache  cache.Cache
	blist  blacklist.BlackLister
}

func newDNSHandler(ctx context.Context, logger logging.Logger,
	settings settings) dns.Handler {
	return &handler{
		dohServers: settings.dohServers,
		ctx:        ctx,
		logger:     logger,
		dial:       newDoHDial(settings),
		client:     &dns.Client{},
		cache:      cache.New(settings.cacheType, settings.cacheOptions...),
		blist: blacklist.NewMap(
			settings.blacklist.fqdnHostnames, settings.blacklist.ips),
	}
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if response := h.cache.Get(r); response != nil {
		response.SetReply(r)
		if err := w.WriteMsg(response); err != nil {
			h.logger.Warn("cannot write DNS message back to client: %s", err)
		}
	}

	if h.blist.FilterRequest(r) {
		response := new(dns.Msg).SetRcode(r, dns.RcodeRefused)
		if err := w.WriteMsg(response); err != nil {
			h.logger.Warn("cannot write DNS message back to client: %s", err)
		}
	}

	DoHConn, err := h.dial(h.ctx, "", "")
	if err != nil {
		h.logger.Warn("cannot dial: %s", err)
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}
	conn := &dns.Conn{Conn: DoHConn}

	response, _, err := h.client.ExchangeWithConn(r, conn)

	if err := conn.Close(); err != nil {
		h.logger.Warn("cannot close the DoT connection: %s", err)
	}

	if err != nil {
		h.logger.Warn("cannot exchange over DoH connection: %s", err)
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}

	if h.blist.FilterResponse(response) {
		response := new(dns.Msg).SetRcode(r, dns.RcodeRefused)
		if err := w.WriteMsg(response); err != nil {
			h.logger.Warn("cannot write DNS message back to client: %s", err)
		}
	}

	h.cache.Add(r, response)

	response.SetReply(r)
	if err := w.WriteMsg(response); err != nil {
		h.logger.Warn("cannot write DNS message back to client: %s", err)
	}
}
