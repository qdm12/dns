package server

import (
	"context"

	"github.com/miekg/dns"
)

var _ dns.Handler = (*handler)(nil)

type handler struct {
	ctx       context.Context //nolint:containedctx
	exchanger exchangerIntf
	warner    Warner
}

func newHandler(ctx context.Context, exchanger exchangerIntf,
	warner Warner,
) *handler {
	return &handler{
		ctx:       ctx,
		exchanger: exchanger,
		warner:    warner,
	}
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	response, err := h.exchanger.Exchange(h.ctx, "udp", r) // note 'udp' is ignored for DoT and DoH
	if err != nil {
		h.warner.Warn(err.Error())
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}

	if response.Truncated { // only happens for UDP plain queries
		response, err = h.exchanger.Exchange(h.ctx, "tcp", r)
		if err != nil {
			h.warner.Warn(err.Error())
			_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
			return
		}
	}

	response.SetReply(r)
	_ = w.WriteMsg(response)
}
