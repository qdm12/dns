package dnssec

import (
	"fmt"

	"github.com/miekg/dns"
)

type dnssecResponse struct {
	answerRRSets    []dnssecRRSet
	authorityRRSets []dnssecRRSet
	rcode           int
}

func (d dnssecResponse) isNXDomain() bool {
	return d.rcode == dns.RcodeNameError
}

func (d dnssecResponse) isNoData() bool {
	return d.rcode == dns.RcodeSuccess && len(d.answerRRSets) == 0
}

func (d dnssecResponse) isSigned() bool {
	// Note a slice of DNSSEC RRSets is either all signed or all unsigned.
	switch {
	case len(d.answerRRSets) > 0 && len(d.answerRRSets[0].rrSigs) == 0,
		len(d.authorityRRSets) > 0 && len(d.authorityRRSets[0].rrSigs) == 0,
		len(d.answerRRSets) == 0 && len(d.authorityRRSets) == 0:
		return false
	default:
		return true
	}
}

func (d dnssecResponse) onlyAnswerRRSet() (rrSet []dns.RR) {
	if len(d.answerRRSets) != 1 {
		panic(fmt.Sprintf("DNSSEC response has %d answer RRSets instead of 1",
			len(d.answerRRSets)))
	}
	return d.answerRRSets[0].rrSet
}

func (d dnssecResponse) onlyAnswerRRSigs() (rrSigs []*dns.RRSIG) {
	if len(d.answerRRSets) != 1 {
		panic(fmt.Sprintf("DNSSEC response has %d answer RRSets instead of 1",
			len(d.answerRRSets)))
	}
	return d.answerRRSets[0].rrSigs
}

func (d dnssecResponse) ToDNSMsg(request *dns.Msg) (response *dns.Msg) {
	response = new(dns.Msg)
	response.SetRcode(request, d.rcode)
	var ignoreTypes []uint16
	if !isRequestAskingForDNSSEC(request) {
		ignoreTypes = []uint16{dns.TypeNSEC, dns.TypeNSEC3, dns.TypeRRSIG}
	}
	response.Answer = dnssecRRSetsToRRs(d.answerRRSets, ignoreTypes...)
	response.Ns = dnssecRRSetsToRRs(d.authorityRRSets, ignoreTypes...)
	return response
}
