package console

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

func (f *Formatter) Response(response *dns.Msg) (s string) {
	s = "id: " + fmt.Sprint(response.Id)

	answerStrings := make([]string, len(response.Answer))
	for i, answer := range response.Answer {
		answerStrings[i] = rrToString(answer)
	}

	switch len(answerStrings) {
	case 0:
		s += "; no answer"
	case 1:
		s += "; answer: " + rrToString(response.Answer[0])
	default:
		s += "; answers: [\n  " +
			strings.Join(answerStrings, ",\n  ") + "\n]"
	}

	// Cache string for calls to RequestResponse
	f.idToResponseString[response.Id] = s

	return s
}

func rrToString(answer dns.RR) (responseStr string) {
	responseStr = answer.String()
	responseStr = strings.TrimPrefix(responseStr, "\t")
	responseStr = strings.ReplaceAll(responseStr, "\t", " ")
	return responseStr
}
