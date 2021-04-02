package lru

import (
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func newTestMsgs(name string) (request, response *dns.Msg) {
	request = &dns.Msg{Question: []dns.Question{{Name: name}}}
	response = &dns.Msg{Answer: []dns.RR{&dns.TXT{Txt: []string{name}}}}
	response = response.Copy() // transform nil slices -> empty slices
	return request, response
}

func Test_lru_e2e(t *testing.T) {
	t.Parallel()

	const (
		maxEntries = 2
		ttl        = time.Hour
	)
	settings := Settings{
		MaxEntries: maxEntries,
		TTL:        ttl,
	}

	requestA, responseA := newTestMsgs("A")
	requestB, responseB := newTestMsgs("B")
	requestC, responseC := newTestMsgs("C")

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
