package dnssec

import (
	"errors"
	"fmt"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/local"
	"github.com/qdm12/dns/v2/internal/stateful"
)

var (
	ErrQuestionsMultiple = errors.New("multiple questions")
)

func Validate(request *dns.Msg, handler dns.Handler) (response *dns.Msg, err error) {
	switch len(request.Question) {
	case 0:
		response = new(dns.Msg)
		response.SetRcode(request, dns.RcodeSuccess)
		return response, nil
	case 1:
	default:
		return nil, fmt.Errorf("%w: %d", ErrQuestionsMultiple, len(request.Question))
	}

	desiredZone := request.Question[0].Name
	qType := request.Question[0].Qtype
	qClass := request.Question[0].Qclass

	if local.IsFQDNLocal(desiredZone) {
		// Do not perform DNSSEC validation for local zones
		writer := stateful.NewWriter()
		handler.ServeDNS(writer, request)
		return writer.Response, nil
	}

	desiredResponse, err := queryRRSets(handler, desiredZone, qClass, qType)
	if err != nil {
		return nil, fmt.Errorf("running desired query: %w", err)
	}

	originalDesiredZone := desiredZone
	cnameTarget := getCnameTarget(desiredResponse.answerRRSets)
	if cnameTarget != "" {
		desiredZone = cnameTarget
	}

	delegationChain, err := buildDelegationChain(handler, desiredZone, qClass)
	if err != nil {
		return nil, fmt.Errorf("building delegation chain for %s: %w",
			originalDesiredZone, err)
	}

	err = validateWithChain(desiredZone, qType, desiredResponse, delegationChain)
	if err != nil {
		return nil, fmt.Errorf("for %s: validating answer RRSets"+
			" with delegation chain: %w",
			nameClassTypeToString(originalDesiredZone, qClass, qType), err)
	}

	return desiredResponse.toDNSMsg(request), nil
}
