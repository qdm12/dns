package lru

import (
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func newTestMsgs(name string, expUnix uint32) (request, response *dns.Msg) {
	request = &dns.Msg{Question: []dns.Question{{Name: name}}}
	response = &dns.Msg{Answer: []dns.RR{&dns.TXT{
		Txt: []string{name},
		Hdr: dns.RR_Header{Ttl: expUnix},
	}}}
	response = response.Copy() // transform nil slices -> empty slices
	return request, response
}

func Test_lru_e2e(t *testing.T) {
	t.Parallel()

	nowUnix := uint32(time.Now().Unix())
	expUnix := nowUnix + 1000

	const (
		maxEntries = 2
	)
	settings := Settings{
		MaxEntries: maxEntries,
	}

	requestA, responseA := newTestMsgs("A", expUnix)
	requestB, responseB := newTestMsgs("B", expUnix)
	requestC, responseC := newTestMsgs("C", expUnix)

	lru := New(settings)

	lru.Add(requestA, responseA)
	lru.Add(requestB, responseB)
	lru.Add(requestA, responseA)
	lru.Add(requestC, responseC)

	response := lru.Get(requestA)
	assert.Equal(t, responseA, response)

	response = lru.Get(requestB)
	assert.Nil(t, response)

	response = lru.Get(requestC)
	assert.Equal(t, responseC, response)
}
