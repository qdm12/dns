package lru

import "github.com/miekg/dns"

func isEmpty(request, response *dns.Msg) (empty bool) {
	if len(request.Question) == 0 || len(response.Answer) == 0 {
		return true
	}

	allEmptyRecords := true
	for _, rr := range response.Answer {
		rrType := rr.Header().Rrtype
		switch rrType {
		// TODO add more DNS record types
		case dns.TypeA:
			record := rr.(*dns.A)
			if record.A != nil {
				allEmptyRecords = false
			}
		case dns.TypeAAAA:
			record := rr.(*dns.AAAA)
			if record.AAAA != nil {
				allEmptyRecords = false
			}
		case dns.TypeTXT:
			record := rr.(*dns.TXT)
			for _, txt := range record.Txt {
				if txt != "" {
					allEmptyRecords = false
					break
				}
			}
		}

		if !allEmptyRecords {
			break
		}
	}

	return allEmptyRecords
}
