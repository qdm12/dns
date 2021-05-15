package log

import "github.com/miekg/dns"

func requestString(request *dns.Msg) (s string) {
	if len(request.Question) == 0 {
		return "[empty request]"
	}
	question := request.Question[0]
	return question.Name + " " +
		dns.Class(question.Qclass).String() + " " +
		dns.Type(question.Qtype).String()
}
