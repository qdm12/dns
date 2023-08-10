package server

import (
	"context"

	"github.com/miekg/dns"
)

var _ dns.Handler = (*Handler)(nil)

type Handler struct {
	ctx      context.Context //nolint:containedctx
	exchange Exchange
	logger   Logger
}

func New(ctx context.Context, exchange Exchange,
	logger Logger) *Handler {
	return &Handler{
		ctx:      ctx,
		exchange: exchange,
		logger:   logger,
	}
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	response, err := h.exchange(h.ctx, r)
	if err != nil {
		h.logger.Warn(err.Error())
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}

	response.SetReply(r)
	_ = w.WriteMsg(response)
}
