package lru

import (
	"strconv"

	"github.com/miekg/dns"
)

type entry struct {
	key      string // from the DNS request
	expUnix  int64  // from the DNS response
	response *dns.Msg
}

func makeKey(request *dns.Msg) (key string) {
	question := request.Question[0]
	key = question.Name + "|" + strconv.Itoa(int(question.Qtype)) + "|" + strconv.Itoa(int(question.Qclass))
	return key
}

func getExpUnix(response *dns.Msg, nowUnix int64) (expUnix int64) {
	secondsLeft := ^uint32(0)
	for _, rr := range response.Answer {
		ttl := rr.Header().Ttl
		if ttl < secondsLeft {
			secondsLeft = ttl
		}
	}
	return nowUnix + int64(secondsLeft)
}
