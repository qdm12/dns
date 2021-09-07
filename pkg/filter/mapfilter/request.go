package mapfilter

import "github.com/miekg/dns"

func (m *Filter) FilterRequest(request *dns.Msg) (blocked bool) {
	for _, question := range request.Question {
		fqdnHostname := question.Name
		_, blocked = m.fqdnHostnames[fqdnHostname]
		if blocked {
			class := dns.ClassToString[question.Qclass]
			qType := dns.TypeToString[question.Qtype]
			m.metrics.HostnamesFilteredInc(class, qType)
			return blocked
		}
	}
	return false
}
