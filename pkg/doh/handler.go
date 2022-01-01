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
	exchange exchangeFunc
	cache    cache.Interface
	filter   filter.Interface
}

func newDNSHandler(ctx context.Context, settings ServerSettings) dns.Handler {
	client := &dns.Client{}
	dial := newDoHDial(settings.Resolver)
	exchange := makeDNSExchange(client, dial, settings.Logger)

	return &handler{
		ctx:      ctx,
		logger:   settings.Logger,
		exchange: exchange,
		cache:    settings.Cache,
		filter:   settings.Filter,
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
		_ = w.WriteMsg(response)
		return
	}

	response, err := h.exchange(h.ctx, r)
	if err != nil {
		h.logger.Warn(err.Error())
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
