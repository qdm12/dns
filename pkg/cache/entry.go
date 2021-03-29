package cache

import (
	"strconv"

	"github.com/miekg/dns"
)

type entry struct {
	key          string // from the DNS request
	creationNano int64
	response     *dns.Msg
}

func makeKey(request *dns.Msg) (key string) {
	question := request.Question[0]
	key = question.Name + "|" + strconv.Itoa(int(question.Qtype)) + "|" + strconv.Itoa(int(question.Qclass))
	return key
}
