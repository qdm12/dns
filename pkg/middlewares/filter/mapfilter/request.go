package mapfilter

import (
	"strings"

	"github.com/miekg/dns"
)

func (m *Filter) FilterRequest(request *dns.Msg) (blocked bool) {
	m.updateLock.RLock()
	defer m.updateLock.RUnlock()

	for _, question := range request.Question {
		fqdnHostname := question.Name
		labels := dns.SplitDomainName(fqdnHostname)

		// Check from least specific to most specific if it is blocked.
		// Root domain '' corresponds to `nil` labels and is always allowed.
		// Single label domains (i.e. localhost) are checked as well.
		for i := range labels {
			labelStartIndex := len(labels) - i - 1
			parent := strings.Join(labels[labelStartIndex:], ".")
			fqdnParent := dns.Fqdn(parent)
			_, blocked = m.fqdnHostnames[fqdnParent]
			if !blocked {
				continue
			}
			class := dns.ClassToString[question.Qclass]
			qType := dns.TypeToString[question.Qtype]
			m.metrics.HostnamesFilteredInc(class, qType)
			return blocked
		}
	}

	return false
}
