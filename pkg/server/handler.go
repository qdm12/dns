package server

import (
	"context"
	"strings"

	"github.com/miekg/dns"
)

var _ dns.Handler = (*handler)(nil)

type handler struct {
	ctx         context.Context //nolint:containedctx
	exchanger   exchangerIntf
	logger      Logger
	timeoutWarn bool
}

func newHandler(ctx context.Context, exchanger exchangerIntf,
	logger Logger, timeoutWarn bool,
) *handler {
	return &handler{
		ctx:         ctx,
		exchanger:   exchanger,
		logger:      logger,
		timeoutWarn: timeoutWarn,
	}
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	response, err := h.exchanger.Exchange(h.ctx, "udp", r) // note 'udp' is ignored for DoT and DoH
	if err != nil {
		h.logErr(err)
		_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
		return
	}

	if response.Truncated { // only happens for UDP plain queries
		response, err = h.exchanger.Exchange(h.ctx, "tcp", r)
		if err != nil {
			h.logErr(err)
			_ = w.WriteMsg(new(dns.Msg).SetRcode(r, dns.RcodeServerFailure))
			return
		}
	}

	response.SetReply(r)
	_ = w.WriteMsg(response)
}

func (h *handler) logErr(err error) {
	if !h.timeoutWarn && strings.HasSuffix(err.Error(), ": i/o timeout") {
		h.logger.Debug(err.Error())
		return
	}
	h.logger.Warn(err.Error())
}
