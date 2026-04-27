package lru

import (
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type entry struct {
	key      string // from the DNS request
	expUnix  int64  // from the DNS response
	response *dns.Msg
}

func makeKey(request *dns.Msg) (key string) {
	question := request.Question[0]
	key = strings.ToLower(question.Name) + "|" +
		strconv.FormatUint(uint64(question.Qtype), 10) + "|" +
		strconv.FormatUint(uint64(question.Qclass), 10)
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
