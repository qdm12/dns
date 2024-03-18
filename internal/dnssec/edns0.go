package dnssec

import "github.com/miekg/dns"

func newEDNSRequest(zone string, qClass, qType uint16) (request *dns.Msg) {
	request = new(dns.Msg).SetQuestion(zone, qType)
	request.Question[0].Qclass = qClass
	request.RecursionDesired = true
	const maxUDPSize = 4096
	const doEdns0 = true
	request.SetEdns0(maxUDPSize, doEdns0)
	return request
}

func isRequestAskingForDNSSEC(request *dns.Msg) bool {
	opt := request.IsEdns0()
	if opt == nil {
		return false
	}

	// See https://datatracker.ietf.org/doc/html/rfc6891#section-6.2.3
	const minUDPSize = 512

	return opt.Hdr.Name == "." &&
		opt.Hdr.Rrtype == dns.TypeOPT &&
		opt.Hdr.Class >= minUDPSize && // UDP size
		opt.Do()
}
