package mapfilter

import (
	"net"
	"net/netip"
	"sync"
	"testing"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/middlewares/filter/update"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Filter(t *testing.T) {
	t.Parallel()

	settings := Settings{
		Update: update.Settings{
			IPs: []netip.Addr{
				netip.AddrFrom4([4]byte{2, 2, 2, 2}),
				netip.AddrFrom4([4]byte{3, 3, 3, 3}),
			},
		},
	}

	settings.Update.BlockHostnames([]string{"github.com", "google.com"})

	filter, err := New(settings)
	require.NoError(t, err)

	assert.True(t, filter.FilterRequest(&dns.Msg{
		Question: []dns.Question{
			{Name: "google.com."},
		},
	}))
	assert.False(t, filter.FilterRequest(&dns.Msg{
		Question: []dns.Question{
			{Name: "duckduckgo.com."},
		},
	}))

	assert.True(t, filter.FilterResponse(&dns.Msg{
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
	assert.False(t, filter.FilterResponse(&dns.Msg{
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{Rrtype: dns.TypeA},
				A:   net.IP{7, 6, 5, 4},
			},
		},
	}))
}

func Test_Filter_threadSafety(t *testing.T) {
	t.Parallel()

	settings := Settings{
		Update: update.Settings{
			IPs: []netip.Addr{
				netip.AddrFrom4([4]byte{2, 2, 2, 2}),
				netip.AddrFrom4([4]byte{3, 3, 3, 3}),
			},
			FqdnHostnames: []string{"github.com."},
		},
	}

	request := &dns.Msg{Question: []dns.Question{
		{Name: "google.com."},
	}}
	response := &dns.Msg{
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{Rrtype: dns.TypeA},
				A:   net.IP{2, 2, 2, 2},
			},
		},
	}

	filter, err := New(settings)
	require.NoError(t, err)

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
			_ = filter.FilterRequest(request)
			_ = filter.FilterResponse(response)
		}()
	}

	endWg.Wait()
}
