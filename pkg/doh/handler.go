package doh

import (
	"context"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/filter"
	"github.com/qdm12/dns/pkg/log"
)

type handler struct {
	// External objects
	ctx    context.Context
	logger log.Logger

	// Internal objects
	dial   dialFunc
	client *dns.Client
	cache  cache.Interface
	filter filter.Interface
}

func newDNSHandler(ctx context.Context, settings ServerSettings) dns.Handler {
	return &handler{
		ctx:    ctx,
		logger: settings.Logger,
		dial:   newDoHDial(settings.Resolver),
		client: &dns.Client{},
		cache:  settings.Cache,
		filter: settings.Filter,
	}
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if h.filter.FilterRequest(r) {
		response := new(dns.Msg).SetRcode(r, dns.RcodeRefused)
		_ = w.WriteMsg(response)
		return
	}

	if response := h.cache.Get(r); response != nil {
		response.SetReply(r)
		if err := w.WriteMsg(response); err != nil {
			h.logger.Warn("cannot write DNS message back to client: " + err.Error())
		}
		return
	}

	DoHConn, err := h.dial(h.ctx, "", "")
	if err != nil {
		h.logger.Warn("cannot dial: " + err.Error())
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}
	conn := &dns.Conn{Conn: DoHConn}

	response, _, err := h.client.ExchangeWithConn(r, conn)

	if closeErr := conn.Close(); closeErr != nil {
		h.logger.Warn("cannot close the DoT connection: " + closeErr.Error())
	}

	if err != nil {
		_ = conn.Close()
		h.logger.Warn("cannot exchange over DoH connection: " + err.Error())
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}

	if h.filter.FilterResponse(response) {
		response := new(dns.Msg).SetRcode(r, dns.RcodeRefused)
		_ = w.WriteMsg(response)
		return
	}

	h.cache.Add(r, response)

	response.SetReply(r)
	_ = w.WriteMsg(response)
}
