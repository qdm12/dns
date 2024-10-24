package server

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_newHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := NewMockLogger(nil)
	exchanger := NewMockexchangerIntf(nil)

	h := newHandler(ctx, exchanger, logger)

	expectedHandler := &handler{
		ctx:       ctx,
		exchanger: exchanger,
		warner:    logger,
	}
	assert.Equal(t, expectedHandler, h)
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

				ctx := context.Background()

				exchanger := NewMockexchangerIntf(ctrl)
				exchanger.EXPECT().Exchange(ctx, "udp", expectedRequest).
					Return(nil, errors.New("test error"))

				logger := NewMockLogger(ctrl)
				logger.EXPECT().Warn("test error")

				return &handler{
					ctx:       ctx,
					exchanger: exchanger,
					warner:    logger,
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
		"exchanged_response": {
			makeHandler: func(t *testing.T, ctrl *gomock.Controller) *handler {
				t.Helper()
				expectedRequest := &dns.Msg{
					Question: []dns.Question{{Name: "test"}},
				}

				ctx := context.Background()

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
