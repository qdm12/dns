package blacklist

import (
	"net"
	"sync"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"inet.af/netaddr"
)

func Test_mapBased(t *testing.T) {
	t.Parallel()

	settings := Settings{
		IPs: []netaddr.IP{
			netaddr.IPv4(2, 2, 2, 2),
			netaddr.IPv4(3, 3, 3, 3),
		},
	}

	settings.BlockHostnames([]string{"github.com", "google.com"})

	blacklister := NewMap(settings)

	assert.True(t, blacklister.FilterRequest(&dns.Msg{
		Question: []dns.Question{
			{Name: "google.com."},
		},
	}))
	assert.False(t, blacklister.FilterRequest(&dns.Msg{
		Question: []dns.Question{
			{Name: "duckduckgo.com."},
		},
	}))

	assert.True(t, blacklister.FilterResponse(&dns.Msg{
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{Rrtype: dns.TypeA},
				A:   net.IP{7, 6, 5, 4},
			},
			&dns.A{
				Hdr: dns.RR_Header{Rrtype: dns.TypeA},
				A:   net.IP{3, 3, 3, 3},
			},
		},
	}))
	assert.False(t, blacklister.FilterResponse(&dns.Msg{
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{Rrtype: dns.TypeA},
				A:   net.IP{7, 6, 5, 4},
			},
		},
	}))
}

func Test_mapBased_threadSafety(t *testing.T) {
	t.Parallel()

	settings := Settings{
		IPs: []netaddr.IP{
			netaddr.IPv4(2, 2, 2, 2),
			netaddr.IPv4(3, 3, 3, 3),
		},
		FqdnHostnames: []string{"github.com."},
	}

	request := &dns.Msg{Question: []dns.Question{
		{Name: "google.com."},
	}}
	response := &dns.Msg{
		Answer: []dns.RR{&dns.A{
			Hdr: dns.RR_Header{Rrtype: dns.TypeA},
			A:   net.IP{2, 2, 2, 2},
		},
		}}

	blacklister := NewMap(settings)

	startWg := new(sync.WaitGroup)
	endWg := new(sync.WaitGroup)

	const parallelism = 1000
	startWg.Add(parallelism)
	endWg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func() {
			defer endWg.Done()
			startWg.Done()
			startWg.Wait()
			_ = blacklister.FilterRequest(request)
			_ = blacklister.FilterResponse(response)
		}()
	}

	endWg.Wait()
}
