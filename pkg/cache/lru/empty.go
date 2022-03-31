package lru

import "github.com/miekg/dns"

func isEmpty(request, response *dns.Msg) (empty bool) {
	if len(request.Question) == 0 || len(response.Answer) == 0 {
		return true
	}

	for _, rr := range response.Answer {
		if !isRREmpty(rr) {
			return false
		}
	}
	return true
}

func isRREmpty(rr dns.RR) (empty bool) {
	rrType := rr.Header().Rrtype
	switch rrType {
	// TODO add more DNS record types
	case dns.TypeA:
		return isAEmpty(rr)
	case dns.TypeAAAA:
		return isAAAAEmpty(rr)
	case dns.TypeTXT:
		return isTXTEmpty(rr)
	default:
		return false
	}
}

func isAEmpty(rr dns.RR) (empty bool) {
	record := rr.(*dns.A) //nolint:forcetypeassert
	return record.A == nil
}

func isAAAAEmpty(rr dns.RR) (empty bool) {
	record := rr.(*dns.AAAA) //nolint:forcetypeassert
	return record.AAAA == nil
}

func isTXTEmpty(rr dns.RR) (empty bool) {
	record := rr.(*dns.TXT) //nolint:forcetypeassert
	for _, txt := range record.Txt {
		if txt != "" {
			return false
		}
	}
	return true
}
