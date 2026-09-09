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

func Test_Filter_FilterResponse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		question        string
		resolvedIP      netip.Addr
		allowedHosts    []string
		blockedIPs      []netip.Addr
		blockedPrefixes []netip.Prefix
		expectedBlocked bool
	}{
		"blocked_ip_not_allowed": {
			question:        "example.com.",
			resolvedIP:      netip.AddrFrom4([4]byte{3, 3, 3, 3}),
			blockedIPs:      []netip.Addr{netip.AddrFrom4([4]byte{3, 3, 3, 3})},
			expectedBlocked: true,
		},
		"blocked_ip_allowed": {
			question:     "example.com.",
			resolvedIP:   netip.AddrFrom4([4]byte{3, 3, 3, 3}),
			allowedHosts: []string{"example.com"},
			blockedIPs:   []netip.Addr{netip.AddrFrom4([4]byte{3, 3, 3, 3})},
		},
		"blocked_ip_subdomain_allowed": {
			question:     "sub.example.com.",
			resolvedIP:   netip.AddrFrom4([4]byte{3, 3, 3, 3}),
			allowedHosts: []string{"example.com"},
			blockedIPs:   []netip.Addr{netip.AddrFrom4([4]byte{3, 3, 3, 3})},
		},
		"blocked_ip_allowed_hostname_different": {
			question:        "sub.other.com.",
			resolvedIP:      netip.AddrFrom4([4]byte{3, 3, 3, 3}),
			allowedHosts:    []string{"example.com"},
			blockedIPs:      []netip.Addr{netip.AddrFrom4([4]byte{3, 3, 3, 3})},
			expectedBlocked: true,
		},
		"blocked_ip_allowed_case_insensitive": {
			question:     "Example.Com.",
			resolvedIP:   netip.AddrFrom4([4]byte{3, 3, 3, 3}),
			allowedHosts: []string{"example.com"},
			blockedIPs:   []netip.Addr{netip.AddrFrom4([4]byte{3, 3, 3, 3})},
		},
		"blocked_ipv6_allowed": {
			question:     "example.com.",
			resolvedIP:   netip.MustParseAddr("2001:db8::1"),
			allowedHosts: []string{"example.com"},
			blockedIPs:   []netip.Addr{netip.MustParseAddr("2001:db8::1")},
		},
		"blocked_prefix_not_allowed": {
			question:        "example.com.",
			resolvedIP:      netip.AddrFrom4([4]byte{3, 3, 3, 3}),
			blockedPrefixes: []netip.Prefix{netip.MustParsePrefix("3.3.3.0/24")},
			expectedBlocked: true,
		},
		"blocked_prefix_allowed": {
			question:     "example.com.",
			resolvedIP:   netip.AddrFrom4([4]byte{3, 3, 3, 3}),
			allowedHosts: []string{"example.com"},
			blockedPrefixes: []netip.Prefix{
				netip.MustParsePrefix("3.3.3.0/24"),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			settings := Settings{
				Update: update.Settings{
					IPs:        testCase.blockedIPs,
					IPPrefixes: testCase.blockedPrefixes,
				},
			}
			settings.Update.SetAllowedHostnames(testCase.allowedHosts)

			filter, err := New(settings)
			require.NoError(t, err)

			var answer dns.RR
			if testCase.resolvedIP.Is4() {
				answer = &dns.A{
					Hdr: dns.RR_Header{Rrtype: dns.TypeA},
					A:   net.IP(testCase.resolvedIP.AsSlice()),
				}
			} else {
				answer = &dns.AAAA{
					Hdr:  dns.RR_Header{Rrtype: dns.TypeAAAA},
					AAAA: net.IP(testCase.resolvedIP.AsSlice()),
				}
			}
			response := &dns.Msg{
				Question: []dns.Question{{Name: testCase.question}},
				Answer:   []dns.RR{answer},
			}

			blocked := filter.FilterResponse(response)

			assert.Equal(t, testCase.expectedBlocked, blocked)
		})
	}
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
	for range parallelism {
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
