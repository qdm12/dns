package lru

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func newTestMsgs(name string, expUnix uint32) (request, response *dns.Msg) {
	request = &dns.Msg{Question: []dns.Question{{Name: name}}}
	response = &dns.Msg{Answer: []dns.RR{&dns.TXT{
		Txt: []string{name},
		Hdr: dns.RR_Header{
			Rrtype: dns.TypeTXT,
			Ttl:    expUnix,
		},
	}}}
	response = response.Copy() // transform nil slices -> empty slices
	return request, response
}

//go:generate mockgen -destination=mock_metrics_test.go -package $GOPACKAGE -mock_names Interface=MockMetrics github.com/qdm12/dns/v2/pkg/cache/metrics Interface

func Test_lru_e2e(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	nowUnix := uint32(time.Now().Unix())
	expUnix := nowUnix + 1000

	const maxEntries = 2
	metrics := NewMockMetrics(ctrl)
	settings := Settings{
		MaxEntries: maxEntries,
		Metrics:    metrics,
	}

	requestA, responseA := newTestMsgs("A", expUnix)
	requestB, responseB := newTestMsgs("B", expUnix)
	requestC, responseC := newTestMsgs("C", expUnix)

	metrics.EXPECT().SetCacheType(CacheType)
	metrics.EXPECT().CacheMaxEntriesSet(settings.MaxEntries)
	lru := New(settings)

	metrics.EXPECT().CacheInsertInc()
	lru.Add(requestA, responseA)

	metrics.EXPECT().CacheInsertInc()
	lru.Add(requestB, responseB)

	metrics.EXPECT().CacheMoveInc()
	lru.Add(requestA, responseA)

	metrics.EXPECT().CacheRemoveInc()
	metrics.EXPECT().CacheInsertInc()
	lru.Add(requestC, responseC)

	metrics.EXPECT().CacheHitInc()
	response := lru.Get(requestA)
	assert.Equal(t, responseA, response)

	metrics.EXPECT().CacheMissInc()
	response = lru.Get(requestB)
	assert.Nil(t, response)

	metrics.EXPECT().CacheHitInc()
	response = lru.Get(requestC)
	assert.Equal(t, responseC, response)
}
