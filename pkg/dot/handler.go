package dot

import (
	"context"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/dns/pkg/cache"
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
	blist  blacklist.BlackLister
}

func newDNSHandler(ctx context.Context, settings ServerSettings) dns.Handler {
	return &handler{
		ctx:    ctx,
		logger: settings.Logger,
		dial:   newDoTDial(settings.Resolver),
		client: &dns.Client{},
		cache:  settings.Cache,
		blist:  settings.Blacklister,
	}
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if h.cache != nil {
		if response := h.cache.Get(r); response != nil {
			response.SetReply(r)
			_ = w.WriteMsg(response)
			return
		}
	}

	if h.blist.FilterRequest(r) {
		response := new(dns.Msg).SetRcode(r, dns.RcodeRefused)
		_ = w.WriteMsg(response)
		return
	}

	DoTConn, err := h.dial(h.ctx, "", "")
	if err != nil {
		h.logger.Warn("cannot dial: " + err.Error())
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}
	conn := &dns.Conn{Conn: DoTConn}

	response, _, err := h.client.ExchangeWithConn(r, conn)

	if closeErr := conn.Close(); closeErr != nil {
		h.logger.Warn("cannot close the DoT connection: " + closeErr.Error())
	}

	if err != nil {
		h.logger.Warn("cannot exchange over DoT connection: " + err.Error())
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}

	if h.blist.FilterResponse(response) {
		response := new(dns.Msg).SetRcode(r, dns.RcodeRefused)
		_ = w.WriteMsg(response)
		return
	}

	if h.cache != nil {
		h.cache.Add(r, response)
	}

	response.SetReply(r)
	_ = w.WriteMsg(response)
}
