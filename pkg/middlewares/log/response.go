package log

import (
	"strings"

	"github.com/miekg/dns"
)

func responseString(response *dns.Msg) (s string) {
	if response == nil {
		return "[]"
	}

	switch len(response.Answer) {
	case 0:
		return "[]"
	case 1:
		return "[" + rrToString(response.Answer[0]) + "]"
	default:
		ss := make([]string, len(response.Answer))
		for i := range response.Answer {
			ss[i] = rrToString(response.Answer[i])
		}
		return "[\n  " + strings.Join(ss, ",\n  ") + "\n]"
	}
}

func rrToString(answer dns.RR) (responseStr string) {
	responseStr = answer.String()
	responseStr = strings.TrimPrefix(responseStr, "\t")
	responseStr = strings.ReplaceAll(responseStr, "\t", " ")
	return responseStr
}
