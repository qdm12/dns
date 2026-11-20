package server

import (
	"errors"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_newHandler(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	logger := NewMockLogger(nil)
	exchanger := NewMockexchangerIntf(nil)
	const timeoutWarn = true

	h := newHandler(ctx, exchanger, logger, timeoutWarn)

	expectedHandler := &handler{
		ctx:         ctx,
		exchanger:   exchanger,
		logger:      logger,
		timeoutWarn: timeoutWarn,
	}
	assert.Equal(t, expectedHandler, h)
}

// newNotFoundAuthority returns the authority section of a NXDOMAIN
// response, as sent by recursive resolvers such as Cloudflare DNS.
func newNotFoundAuthority() (rrs []dns.RR) {
	return []dns.RR{&dns.NS{
		Hdr: dns.RR_Header{
			Name:   "test.",
			Rrtype: dns.TypeNS,
			Class:  dns.ClassINET,
			Ttl:    300,
		},
		Ns: "ns.test.",
	}}
}

func Test_Handler_ServeDNS(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		makeHandler func(t *testing.T, ctrl *gomock.Controller) *handler
		request     *dns.Msg
		response    *dns.Msg
	}{
		"exchange_error": {
			makeHandler: func(t *testing.T, ctrl *gomock.Controller) *handler {
				t.Helper()
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}

				ctx := t.Context()

				exchanger := NewMockexchangerIntf(ctrl)
				exchanger.EXPECT().Exchange(ctx, "udp", expectedRequest).
					Return(nil, errors.New("test error"))

				logger := NewMockLogger(ctrl)
				logger.EXPECT().Warn("test error")

				return &handler{
					ctx:       ctx,
					exchanger: exchanger,
					logger:    logger,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{{Name: "test"}},
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
					Rcode:    dns.RcodeServerFailure,
				},
				Question: []dns.Question{{Name: "test"}},
			},
		},
		"udp_buffer_too_small_retries_tcp": {
			makeHandler: func(t *testing.T, ctrl *gomock.Controller) *handler {
				t.Helper()
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}

				ctx := t.Context()

				exchanger := NewMockexchangerIntf(ctrl)
				exchanger.EXPECT().Exchange(ctx, "udp", expectedRequest).
					Return(nil, dns.ErrBuf)
				exchanger.EXPECT().Exchange(ctx, "tcp", expectedRequest).
					Return(&dns.Msg{Answer: []dns.RR{&dns.A{}}}, nil)

				return &handler{
					ctx:       ctx,
					exchanger: exchanger,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{{Name: "test"}},
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
				},
				Question: []dns.Question{{Name: "test"}},
				Answer:   []dns.RR{&dns.A{}},
			},
		},
		"exchanged_response": {
			makeHandler: func(t *testing.T, ctrl *gomock.Controller) *handler {
				t.Helper()
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}

				ctx := t.Context()

				exchanger := NewMockexchangerIntf(ctrl)
				exchanger.EXPECT().Exchange(ctx, "udp", expectedRequest).
					Return(&dns.Msg{Answer: []dns.RR{&dns.A{}}}, nil)

				return &handler{
					ctx:       ctx,
					exchanger: exchanger,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{{Name: "test"}},
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
				},
				Question: []dns.Question{{Name: "test"}},
				Answer:   []dns.RR{&dns.A{}},
			},
		},
		// See https://github.com/passteque/gluetun/issues/3552
		"upstream_nxdomain": {
			makeHandler: func(t *testing.T, ctrl *gomock.Controller) *handler {
				t.Helper()
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}

				ctx := t.Context()

				exchanger := NewMockexchangerIntf(ctrl)
				exchanger.EXPECT().Exchange(ctx, "udp", expectedRequest).
					Return(&dns.Msg{
						MsgHdr: dns.MsgHdr{Rcode: dns.RcodeNameError},
						Ns:     newNotFoundAuthority(),
					}, nil)

				return &handler{
					ctx:       ctx,
					exchanger: exchanger,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{{Name: "test"}},
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
					Rcode:    dns.RcodeNameError,
				},
				Question: []dns.Question{{Name: "test"}},
				Ns:       newNotFoundAuthority(),
			},
		},
		// See https://github.com/passteque/gluetun/issues/3552
		"upstream_servfail": {
			makeHandler: func(t *testing.T, ctrl *gomock.Controller) *handler {
				t.Helper()
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}

				ctx := t.Context()

				exchanger := NewMockexchangerIntf(ctrl)
				exchanger.EXPECT().Exchange(ctx, "udp", expectedRequest).
					Return(&dns.Msg{MsgHdr: dns.MsgHdr{Rcode: dns.RcodeServerFailure}}, nil)

				return &handler{
					ctx:       ctx,
					exchanger: exchanger,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{{Name: "test"}},
			},
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
					Rcode:    dns.RcodeServerFailure,
				},
				Question: []dns.Question{{Name: "test"}},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			handler := testCase.makeHandler(t, ctrl)
			writer := &testWriter{}

			handler.ServeDNS(writer, testCase.request)

			assert.Equal(t, testCase.response, writer.responseWritten)
		})
	}
}
