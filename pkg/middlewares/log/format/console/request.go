package console

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

func (f *Formatter) Request(request *dns.Msg) (s string) {
	s = "id: " + fmt.Sprint(request.Id)

	questionStrings := make([]string, len(request.Question))
	for i, question := range request.Question {
		questionStrings[i] = question.Name + " " +
			dns.Class(question.Qclass).String() + " " +
			dns.Type(question.Qtype).String()
	}

	switch len(questionStrings) {
	case 0:
		s += "; no question"
	case 1:
		s += "; question: " + questionStrings[0]
	default:
		s += "; questions: " + strings.Join(questionStrings, ", ")
	}

	// Cache string for calls to RequestResponse
	f.idToRequestString[request.Id] = s

	return s
}
