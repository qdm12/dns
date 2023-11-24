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
	cachemiddleware "github.com/qdm12/dns/v2/pkg/middlewares/cache"
	filtermiddleware "github.com/qdm12/dns/v2/pkg/middlewares/filter"
	metricsmiddleware "github.com/qdm12/dns/v2/pkg/middlewares/metrics"
	"github.com/qdm12/log"
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
	server, err := NewServer(ServerSettings{
		ListeningAddress: ptrTo(""),
		Logger:           log.New(),
	})
	require.NoError(t, err)

	runError, startErr := server.Start()
	require.NoError(t, startErr)

	listeningAddress, err := server.ListeningAddress()
	require.NoError(t, err)

	resolver := &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial: func(ctx context.Context, network string, address string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: time.Second}
			return dialer.DialContext(ctx, "udp", listeningAddress.String())
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

	cache := NewMockcache(ctrl)
	cache.EXPECT().
		Get(mockhelp.NewMatcherRequest(expectedRequestA)).
		Return(nil)
	cache.EXPECT().
		Get(mockhelp.NewMatcherRequest(expectedRequestAAAA)).
		Return(nil)
	cache.EXPECT().Add(
		mockhelp.NewMatcherRequest(expectedRequestA),
		mockhelp.NewMatcherResponse(mockhelp.MatcherResponseSettings{
			Response:           expectedResponseA,
			OnlyHasAnswerTypes: []uint16{dns.TypeA},
			IgnoreAnswerTypes:  []uint16{dns.TypeA},
		}))
	cache.EXPECT().Add(
		mockhelp.NewMatcherRequest(expectedRequestAAAA),
		mockhelp.NewMatcherResponse(mockhelp.MatcherResponseSettings{
			Response:           expectedResponseAAAA,
			OnlyHasAnswerTypes: []uint16{dns.TypeAAAA},
			IgnoreAnswerTypes:  []uint16{dns.TypeAAAA},
		}))
	cacheMiddleware, err := cachemiddleware.New(cachemiddleware.Settings{Cache: cache})
	require.NoError(t, err)

	filter := NewMockfilter(ctrl)
	filter.EXPECT().
		FilterRequest(mockhelp.NewMatcherRequest(expectedRequestA)).
		Return(false)
	filter.EXPECT().
		FilterRequest(mockhelp.NewMatcherRequest(expectedRequestAAAA)).
		Return(false)
	filter.EXPECT().
		FilterResponse(mockhelp.NewMatcherResponse(mockhelp.MatcherResponseSettings{
			Response:           expectedResponseA,
			OnlyHasAnswerTypes: []uint16{dns.TypeA},
			IgnoreAnswerTypes:  []uint16{dns.TypeA},
		})).Return(false)
	filter.EXPECT().
		FilterResponse(mockhelp.NewMatcherResponse(mockhelp.MatcherResponseSettings{
			Response:           expectedResponseAAAA,
			OnlyHasAnswerTypes: []uint16{dns.TypeAAAA},
			IgnoreAnswerTypes:  []uint16{dns.TypeAAAA},
		})).Return(false)
	filterMiddleware, err := filtermiddleware.New(filtermiddleware.Settings{Filter: filter})
	require.NoError(t, err)

	logger := NewMockLogger(ctrl)
	logger.EXPECT().Info(mockhelp.NewMatcherRegex("DNS server listening on .*:[1-9][0-9]{0,4}"))

	metrics := NewMockMetrics(ctrl)
	metrics.EXPECT().
		DoHDialInc("https://cloudflare-dns.com/dns-query").
		Times(2)
	middlewareMetrics := NewMockmiddlewareMetrics(ctrl)
	middlewareMetrics.EXPECT().InFlightRequestsInc().Times(2)
	middlewareMetrics.EXPECT().InFlightRequestsDec().Times(2)
	middlewareMetrics.EXPECT().RequestsInc().Times(2)
	middlewareMetrics.EXPECT().ResponsesInc().Times(2)
	middlewareMetrics.EXPECT().QuestionsInc("IN", "A")
	middlewareMetrics.EXPECT().QuestionsInc("IN", "AAAA")
	middlewareMetrics.EXPECT().RcodeInc("NOERROR").Times(2)
	middlewareMetrics.EXPECT().AnswersInc("IN", "A").MinTimes(1)
	middlewareMetrics.EXPECT().AnswersInc("IN", "AAAA").MinTimes(1)

	metricsMiddleware, err := metricsmiddleware.New(
		metricsmiddleware.Settings{
			Metrics: middlewareMetrics,
		},
	)
	require.NoError(t, err)

	server, err := NewServer(ServerSettings{
		Logger:      logger,
		Middlewares: []Middleware{metricsMiddleware, cacheMiddleware, filterMiddleware},
		Resolver: ResolverSettings{
			Metrics: metrics,
		},
		ListeningAddress: ptrTo(""),
	})
	require.NoError(t, err)

	runError, startErr := server.Start()
	require.NoError(t, startErr)

	listeningAddress, err := server.ListeningAddress()
	require.NoError(t, err)

	resolver := &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial: func(ctx context.Context, network string, address string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: time.Second}
			return dialer.DialContext(ctx, "udp", listeningAddress.String())
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
