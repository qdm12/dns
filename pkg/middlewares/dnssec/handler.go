package dnssec

import (
	"errors"
	"strings"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/dnssec"
)

type handler struct {
	// Injected from middleware
	logger    Logger
	validator *dnssec.Validator
	next      dns.Handler
}

func newHandler(logger Logger, validator *dnssec.Validator, next dns.Handler) *handler {
	return &handler{
		logger:    logger,
		validator: validator,
		next:      next,
	}
}

func (h *handler) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	response, err := h.validator.Validate(request, h.next)
	if err != nil {
		message := err.Error()
		if errors.Is(err, dnssec.ErrRcodeBad) && strings.HasSuffix(message, dns.RcodeToString[dns.RcodeServerFailure]) {
			// Transient server errors from upstream resolvers, log at debug level
			h.logger.Debug(message)
		} else {
			h.logger.Warn(message)
		}
		response = new(dns.Msg)
		response.SetRcode(request, dns.RcodeServerFailure)
	}

	_ = w.WriteMsg(response)
}
