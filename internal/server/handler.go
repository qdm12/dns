package server

import (
	"context"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/cache"
	"github.com/qdm12/dns/v2/pkg/filter"
	"github.com/qdm12/dns/v2/pkg/log"
)

var _ dns.Handler = (*Handler)(nil)

type Handler struct {
	ctx      context.Context //nolint:containedctx
	exchange Exchange
	filter   filter.Interface
	cache    cache.Interface
	logger   log.Logger
}

func New(ctx context.Context, exchange Exchange,
	filter filter.Interface, cache cache.Interface,
	logger log.Logger) *Handler {
	return &Handler{
		ctx:      ctx,
		exchange: exchange,
		filter:   filter,
		cache:    cache,
		logger:   logger,
	}
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
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
