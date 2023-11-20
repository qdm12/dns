//go:build integration
// +build integration

package dot

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/mockhelp"
	cachemiddleware "github.com/qdm12/dns/v2/pkg/middlewares/cache"
	filtermiddleware "github.com/qdm12/dns/v2/pkg/middlewares/filter"
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
	server, err := NewServer(ServerSettings{
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

	const hostname = "google.com" // we use google.com as github.com doesn't have an IPv6 :(
	ips, err := resolver.LookupIPAddr(context.Background(), hostname)

	require.NoError(t, err)
	require.NotEmpty(t, ips)
	t.Logf("resolved %s to: %v", hostname, ips)

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
				Option: []dns.EDNS0{
					&dns.EDNS0_PADDING{
						// padding changes with message content
						// the comparison only checks both length
						// are empty or non-empty.
						Padding: []byte{0},
					},
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
				Option: []dns.EDNS0{
					&dns.EDNS0_PADDING{
						// padding changes with message content
						// the comparison only checks both length
						// are empty or non-empty.
						Padding: []byte{0},
					},
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
		})).MinTimes(1)
	cache.EXPECT().Add(
		mockhelp.NewMatcherRequest(expectedRequestAAAA),
		mockhelp.NewMatcherResponse(mockhelp.MatcherResponseSettings{
			Response:           expectedResponseAAAA,
			OnlyHasAnswerTypes: []uint16{dns.TypeAAAA},
			IgnoreAnswerTypes:  []uint16{dns.TypeAAAA},
		})).MinTimes(1)
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
		})).
		Return(false).MinTimes(1)
	filter.EXPECT().
		FilterResponse(mockhelp.NewMatcherResponse(mockhelp.MatcherResponseSettings{
			Response:           expectedResponseAAAA,
			OnlyHasAnswerTypes: []uint16{dns.TypeAAAA},
			IgnoreAnswerTypes:  []uint16{dns.TypeAAAA},
		})).
		Return(false).MinTimes(1)
	filterMiddleware, err := filtermiddleware.New(filtermiddleware.Settings{Filter: filter})
	require.NoError(t, err)

	logger := NewMockLogger(ctrl)
	logger.EXPECT().Info(mockhelp.NewMatcherRegex("DNS server listening on .*:[1-9][0-9]{0,4}"))

	dotMetrics := NewMockMetrics(ctrl)
	dotMetrics.EXPECT().
		DoTDialInc("cloudflare-dns.com",
			mockhelp.NewMatcherOneOf("1.1.1.1:853", "1.0.0.1:853"), "success").
		Times(2)

	// middleware metrics
	metrics := NewMockmiddlewareMetrics(ctrl)
	metrics.EXPECT().InFlightRequestsInc().Times(2)
	metrics.EXPECT().InFlightRequestsDec().Times(2)
	metrics.EXPECT().RequestsInc().Times(2)
	metrics.EXPECT().ResponsesInc().Times(2)
	metrics.EXPECT().QuestionsInc("IN", "A")
	metrics.EXPECT().QuestionsInc("IN", "AAAA")
	metrics.EXPECT().RcodeInc("NOERROR").Times(2)
	metrics.EXPECT().AnswersInc("IN", "A").MinTimes(1)
	metrics.EXPECT().AnswersInc("IN", "AAAA").MinTimes(1)

	metricsMiddleware, err := metricsmiddleware.New(
		metricsmiddleware.Settings{
			Metrics: metrics,
		},
	)
	require.NoError(t, err)

	server, err := NewServer(ServerSettings{
		Logger:      logger,
		Middlewares: []Middleware{metricsMiddleware, cacheMiddleware, filterMiddleware},
		Resolver: ResolverSettings{
			Metrics: dotMetrics,
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
