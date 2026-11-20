package cache

import (
	"net/netip"
	"testing"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/mockhelp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Middleware_ServeDNS(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		makeMiddleware func(t *testing.T, ctrl *gomock.Controller, request *dns.Msg) (
			middleware *Middleware, next dns.Handler)
		request *dns.Msg
		// Expected response written to the client, whatever the cache
		// held or the next handler wrote.
		expectedResponse *dns.Msg
		// Answer types the expected response must exclusively hold.
		expectedAnswerTypes []uint16
	}{
		"cache_hit_name_error": {
			makeMiddleware: func(t *testing.T, ctrl *gomock.Controller, request *dns.Msg) (
				*Middleware, dns.Handler,
			) {
				t.Helper()

				// A cached response does not carry the reply fields of the
				// request it was stored for, they are set when it is served.
				cachedResponse := &dns.Msg{
					MsgHdr: dns.MsgHdr{Response: true, Rcode: dns.RcodeNameError},
				}

				cacheMock := NewMockCache(ctrl)
				cacheMock.EXPECT().
					Get(mockhelp.NewMatcherRequest(request)).
					Return(cachedResponse)
				// No Add call expected: a cache hit adds nothing to the cache.

				middleware, err := New(Settings{Cache: cacheMock})
				require.NoError(t, err)

				return middleware, dns.HandlerFunc(func(dns.ResponseWriter, *dns.Msg) {
					t.Error("next handler called although the cache holds a response")
				})
			},
			request: newTestRequest(),
			expectedResponse: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Id:               1234,
					Response:         true,
					RecursionDesired: true,
					Rcode:            dns.RcodeNameError,
				},
				Question: newTestQuestion(),
			},
		},
		"cache_hit_answer": {
			makeMiddleware: func(t *testing.T, ctrl *gomock.Controller, request *dns.Msg) (
				*Middleware, dns.Handler,
			) {
				t.Helper()

				cachedResponse := &dns.Msg{
					MsgHdr: dns.MsgHdr{Response: true, Rcode: dns.RcodeSuccess},
					Answer: []dns.RR{newTestAnswer()},
				}

				cacheMock := NewMockCache(ctrl)
				cacheMock.EXPECT().
					Get(mockhelp.NewMatcherRequest(request)).
					Return(cachedResponse)

				middleware, err := New(Settings{Cache: cacheMock})
				require.NoError(t, err)

				return middleware, dns.HandlerFunc(func(dns.ResponseWriter, *dns.Msg) {
					t.Error("next handler called although the cache holds a response")
				})
			},
			request: newTestRequest(),
			expectedResponse: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Id:               1234,
					Response:         true,
					RecursionDesired: true,
					Rcode:            dns.RcodeSuccess,
				},
				Question: newTestQuestion(),
				Answer:   []dns.RR{newTestAnswer()},
			},
			expectedAnswerTypes: []uint16{dns.TypeA},
		},
		"cache_miss_server_failure": {
			makeMiddleware: func(t *testing.T, ctrl *gomock.Controller, request *dns.Msg) (
				*Middleware, dns.Handler,
			) {
				t.Helper()

				serverFailureResponse := new(dns.Msg).SetRcode(request, dns.RcodeServerFailure)

				cacheMock := NewMockCache(ctrl)
				cacheMock.EXPECT().
					Get(mockhelp.NewMatcherRequest(request)).
					Return(nil)
				cacheMock.EXPECT().
					Add(mockhelp.NewMatcherRequest(request), serverFailureResponse)

				middleware, err := New(Settings{Cache: cacheMock})
				require.NoError(t, err)

				// The next handler behaves like the server handler: it sets the
				// reply fields of the response received from upstream, leaving
				// its rcode untouched.
				return middleware, dns.HandlerFunc(func(writer dns.ResponseWriter, _ *dns.Msg) {
					_ = writer.WriteMsg(serverFailureResponse)
				})
			},
			request: newTestRequest(),
			expectedResponse: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Id:               1234,
					Response:         true,
					RecursionDesired: true,
					Rcode:            dns.RcodeServerFailure,
				},
				Question: newTestQuestion(),
			},
		},
		"cache_miss_answer": {
			makeMiddleware: func(t *testing.T, ctrl *gomock.Controller, request *dns.Msg) (
				*Middleware, dns.Handler,
			) {
				t.Helper()

				answerResponse := newTestResponse(request)

				cacheMock := NewMockCache(ctrl)
				cacheMock.EXPECT().
					Get(mockhelp.NewMatcherRequest(request)).
					Return(nil)
				cacheMock.EXPECT().
					Add(mockhelp.NewMatcherRequest(request), answerResponse)

				middleware, err := New(Settings{Cache: cacheMock})
				require.NoError(t, err)

				return middleware, dns.HandlerFunc(func(writer dns.ResponseWriter, _ *dns.Msg) {
					_ = writer.WriteMsg(answerResponse)
				})
			},
			request: newTestRequest(),
			expectedResponse: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Id:               1234,
					Response:         true,
					RecursionDesired: true,
					Rcode:            dns.RcodeSuccess,
				},
				Question: newTestQuestion(),
				Answer:   []dns.RR{newTestAnswer()},
			},
			expectedAnswerTypes: []uint16{dns.TypeA},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			middleware, next := testCase.makeMiddleware(t, ctrl, testCase.request)
			wrappedHandler := middleware.Wrap(next)

			writer := NewMockResponseWriter(ctrl)
			writer.EXPECT().
				WriteMsg(mockhelp.NewMatcherResponse(mockhelp.MatcherResponseSettings{
					Response:           testCase.expectedResponse,
					OnlyHasAnswerTypes: testCase.expectedAnswerTypes,
				})).
				Return(nil)

			wrappedHandler.ServeDNS(writer, testCase.request)
		})
	}
}

func Test_Middleware_String(t *testing.T) {
	t.Parallel()

	// The cache is not used to describe the middleware.
	middleware := &Middleware{}

	assert.Equal(t, "cache", middleware.String())
}

func newTestRequest() *dns.Msg {
	const requestID uint16 = 1234
	return &dns.Msg{
		MsgHdr:   dns.MsgHdr{Id: requestID, RecursionDesired: true},
		Question: newTestQuestion(),
	}
}

func newTestQuestion() []dns.Question {
	return []dns.Question{{
		Name:   "test.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}}
}

func newTestAnswer() *dns.A {
	const answerTTL uint32 = 300
	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   "test.",
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    answerTTL,
		},
		A: netip.MustParseAddr("1.2.3.4").AsSlice(),
	}
}

func newTestResponse(request *dns.Msg) *dns.Msg {
	response := new(dns.Msg)
	response.SetReply(request)
	response.Answer = []dns.RR{newTestAnswer()}
	return response
}
