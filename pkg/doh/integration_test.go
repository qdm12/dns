//go:build integration
// +build integration

package doh

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/mockhelp"
	metricsmiddleware "github.com/qdm12/dns/v2/pkg/middlewares/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Resolver(t *testing.T) {
	t.Parallel()

	const hostname = "google.com"

	resolver, err := NewResolver(ResolverSettings{})
	require.NoError(t, err)

	ips, err := resolver.LookupIPAddr(context.Background(), hostname)

	require.NoError(t, err)
	require.NotEmpty(t, ips)
	t.Logf("resolved %s to: %v", hostname, ips)
}

func Test_Server(t *testing.T) {
	server, err := NewServer(ServerSettings{})
	require.NoError(t, err)

	runError, startErr := server.Start()
	require.NoError(t, startErr)

	resolver := &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial: func(ctx context.Context, network string, address string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: time.Second}
			return dialer.DialContext(ctx, "udp", "127.0.0.1:53")
		},
	}

	const parallelResolutions = 1
	startWg := new(sync.WaitGroup)
	endWg := new(sync.WaitGroup)
	startWg.Add(parallelResolutions)
	endWg.Add(parallelResolutions)
	hostnames := []string{
		"google.com", "google.com", "github.com", "amazon.com", "cloudflare.com",
	}

	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < parallelResolutions; i++ {
		hostnameIndex := i % len(hostnames)
		hostname := hostnames[hostnameIndex]
		go func() {
			startWg.Done()
			startWg.Wait()
			ips, err := resolver.LookupIPAddr(ctx, hostname)
			assert.NoError(t, err)
			assert.NotEmpty(t, ips)
			t.Log(ips)
			endWg.Done()
		}()
	}

	endWg.Wait()
	cancel()

	select {
	case err := <-runError:
		assert.NoError(t, err)
	default:
	}

	err = server.Stop()
	assert.NoError(t, err)
}

//go:generate mockgen -destination=mock_cache_test.go -package $GOPACKAGE -mock_names Interface=MockCache github.com/qdm12/dns/v2/pkg/cache Interface
//go:generate mockgen -destination=mock_doh_metrics_test.go -package $GOPACKAGE -mock_names Interface=MockDoHMetrics github.com/qdm12/dns/v2/pkg/doh/metrics Interface
//go:generate mockgen -destination=mock_filter_test.go -package $GOPACKAGE -mock_names Interface=MockFilter github.com/qdm12/dns/v2/pkg/filter Interface
//go:generate mockgen -destination=mock_logger_test.go -package $GOPACKAGE github.com/qdm12/dns/v2/pkg/log Logger

func Test_Server_Mocks(t *testing.T) {
	ctrl := gomock.NewController(t)

	expectedRequestA := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Opcode:           dns.OpcodeQuery,
			Rcode:            dns.RcodeSuccess,
			RecursionDesired: true,
		},
		Question: []dns.Question{{
			Name:   "google.com.",
			Qtype:  dns.TypeA,
			Qclass: dns.ClassINET,
		}},
		Extra: []dns.RR{
			&dns.OPT{
				Hdr: dns.RR_Header{
					Name:   ".",
					Rrtype: dns.TypeOPT,
					Class:  1232, // UDP size
				},
			},
		},
	}
	expectedResponseA := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Opcode:             dns.OpcodeQuery,
			Rcode:              dns.RcodeSuccess,
			Response:           true,
			RecursionDesired:   true,
			RecursionAvailable: true,
		},
		Question: []dns.Question{{
			Name:   "google.com.",
			Qtype:  dns.TypeA,
			Qclass: dns.ClassINET,
		}},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "google.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
				},
				A: net.IP{1, 2, 3, 4}, // compared on length
			},
		},
		Extra: []dns.RR{
			&dns.OPT{
				Hdr: dns.RR_Header{
					Name:   ".",
					Rrtype: dns.TypeOPT,
					Class:  1232, // UDP size
				},
			},
		},
	}

	expectedRequestAAAA := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Opcode:           dns.OpcodeQuery,
			Rcode:            dns.RcodeSuccess,
			RecursionDesired: true,
		},
		Question: []dns.Question{{
			Name:   "google.com.",
			Qtype:  dns.TypeAAAA,
			Qclass: dns.ClassINET,
		}},
		Extra: []dns.RR{
			&dns.OPT{
				Hdr: dns.RR_Header{
					Name:   ".",
					Rrtype: dns.TypeOPT,
					Class:  1232, // UDP size
				},
			},
		},
	}
	expectedResponseAAAA := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Opcode:             dns.OpcodeQuery,
			Rcode:              dns.RcodeSuccess,
			Response:           true,
			RecursionDesired:   true,
			RecursionAvailable: true,
		},
		Question: []dns.Question{{
			Name:   "google.com.",
			Qtype:  dns.TypeAAAA,
			Qclass: dns.ClassINET,
		}},
		Answer: []dns.RR{
			&dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   "google.com.",
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
				},
				AAAA: net.IP{1, 2, 3, 4}, // compared on length > 0
			},
		},
		Extra: []dns.RR{
			&dns.OPT{
				Hdr: dns.RR_Header{
					Name:   ".",
					Rrtype: dns.TypeOPT,
					Class:  1232, // UDP size
				},
			},
		},
	}

	cache := NewMockCache(ctrl)
	cache.EXPECT().
		Get(mockhelp.NewMatcherRequest(expectedRequestA)).
		Return(nil)
	cache.EXPECT().
		Get(mockhelp.NewMatcherRequest(expectedRequestAAAA)).
		Return(nil)

	cache.EXPECT().Add(
		mockhelp.NewMatcherRequest(expectedRequestA),
		mockhelp.NewMatcherResponse(expectedResponseA))
	cache.EXPECT().Add(
		mockhelp.NewMatcherRequest(expectedRequestAAAA),
		mockhelp.NewMatcherResponse(expectedResponseAAAA))

	filter := NewMockFilter(ctrl)
	filter.EXPECT().
		FilterRequest(mockhelp.NewMatcherRequest(expectedRequestA)).
		Return(false)
	filter.EXPECT().
		FilterRequest(mockhelp.NewMatcherRequest(expectedRequestAAAA)).
		Return(false)
	filter.EXPECT().
		FilterResponse(mockhelp.NewMatcherResponse(expectedResponseA)).
		Return(false)
	filter.EXPECT().
		FilterResponse(mockhelp.NewMatcherResponse(expectedResponseAAAA)).
		Return(false)

	logger := NewMockLogger(ctrl)
	logger.EXPECT().Info("DNS server listening on :53")

	metrics := NewMockDoHMetrics(ctrl)
	metrics.EXPECT().
		DoTDialInc("cloudflare-dns.com",
			mockhelp.NewMatcherOneOf("1.1.1.1:853", "1.0.0.1:853"), "success").
		Times(2)
	metrics.EXPECT().
		DoHDialInc("https://cloudflare-dns.com/dns-query").
		Times(2)
	// middleware metrics
	metrics.EXPECT().InFlightRequestsInc().Times(2)
	metrics.EXPECT().InFlightRequestsDec().Times(2)
	metrics.EXPECT().RequestsInc().Times(2)
	metrics.EXPECT().ResponsesInc().Times(2)
	metrics.EXPECT().QuestionsInc("IN", "A")
	metrics.EXPECT().QuestionsInc("IN", "AAAA")
	metrics.EXPECT().RcodeInc("NOERROR").Times(2)

	metricsMiddleware := metricsmiddleware.New(
		metricsmiddleware.Settings{
			Metrics: metrics,
		},
	)

	server, err := NewServer(ServerSettings{
		Cache:       cache,
		Filter:      filter,
		Logger:      logger,
		Middlewares: []Middleware{metricsMiddleware},
		Resolver: ResolverSettings{
			Metrics: metrics,
		},
	})
	require.NoError(t, err)

	runError, startErr := server.Start()
	require.NoError(t, startErr)

	resolver := &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial: func(ctx context.Context, network string, address string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: time.Second}
			return dialer.DialContext(ctx, "udp", "127.0.0.1:53")
		},
	}

	const hostname = "google.com"
	ips, err := resolver.LookupIPAddr(context.Background(), hostname)
	assert.NoError(t, err)
	assert.NotEmpty(t, ips)
	t.Log(ips)

	select {
	case err := <-runError:
		assert.NoError(t, err)
	default:
	}
	err = server.Stop()
	assert.NoError(t, err)
}
